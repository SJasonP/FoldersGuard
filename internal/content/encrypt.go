package content

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	fgcrypto "foldersguard/internal/crypto"
	"foldersguard/internal/model"
)

const (
	objectMagic      = "FGOBJv1\n"
	noncePrefixSize  = 8
	defaultChunkSize = 4 * 1024 * 1024
)

type FileSource struct {
	FileID       string
	AbsolutePath string
	Key          []byte
	StorageKind  model.StorageKind
	VisiblePath  string
	Parts        []model.Part
}

type Encryptor struct {
	OutputRoot string
	ChunkSize  int
}

func (e Encryptor) EncryptFile(ctx context.Context, source FileSource) error {
	if e.OutputRoot == "" {
		return fmt.Errorf("output root is required")
	}
	if source.AbsolutePath == "" {
		return fmt.Errorf("source path is required")
	}
	if source.VisiblePath == "" {
		return fmt.Errorf("visible path is required")
	}

	aead, err := fgcrypto.NewAES256GCM(source.Key)
	if err != nil {
		return err
	}

	switch source.StorageKind {
	case model.StorageKindSingle:
		return e.encryptSingle(ctx, aead, source)
	case model.StorageKindSplit:
		return e.encryptSplit(ctx, aead, source)
	default:
		return fmt.Errorf("unsupported storage kind %q", source.StorageKind)
	}
}

func (e Encryptor) encryptSingle(ctx context.Context, aead cipher.AEAD, source FileSource) error {
	input, err := os.Open(source.AbsolutePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer input.Close()

	outputPath := filepath.Join(e.OutputRoot, filepath.FromSlash(source.VisiblePath))
	associatedData := []byte("fg-content-v1:file:" + source.FileID)
	if err := sealReader(ctx, aead, input, outputPath, associatedData); err != nil {
		return fmt.Errorf("encrypt single file: %w", err)
	}
	return nil
}

func (e Encryptor) encryptSplit(ctx context.Context, aead cipher.AEAD, source FileSource) error {
	if len(source.Parts) == 0 {
		return fmt.Errorf("split file requires parts")
	}

	input, err := os.Open(source.AbsolutePath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer input.Close()

	dir := filepath.Join(e.OutputRoot, filepath.FromSlash(source.VisiblePath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create split output directory: %w", err)
	}

	for _, part := range source.Parts {
		if err := ctx.Err(); err != nil {
			return err
		}
		section := io.NewSectionReader(input, part.Offset, part.Size)
		outputPath := filepath.Join(dir, part.VisibleName.String())
		associatedData := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", source.FileID, part.Index, part.Offset, part.Size))
		if err := sealReader(ctx, aead, section, outputPath, associatedData); err != nil {
			return fmt.Errorf("encrypt part %d: %w", part.Index, err)
		}
	}
	return nil
}

func sealReader(ctx context.Context, aead cipher.AEAD, reader io.Reader, outputPath string, associatedData []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create ciphertext: %w", err)
	}
	defer output.Close()

	if err := writeEncryptedObject(ctx, aead, reader, output, associatedData, defaultChunkSize); err != nil {
		return err
	}
	return nil
}

func writeEncryptedObject(ctx context.Context, aead cipher.AEAD, reader io.Reader, writer io.Writer, associatedData []byte, chunkSize int) error {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}
	if aead.NonceSize() != noncePrefixSize+4 {
		return fmt.Errorf("unsupported nonce size %d", aead.NonceSize())
	}

	noncePrefix := make([]byte, noncePrefixSize)
	if _, err := rand.Read(noncePrefix); err != nil {
		return fmt.Errorf("generate nonce prefix: %w", err)
	}

	if _, err := writer.Write([]byte(objectMagic)); err != nil {
		return fmt.Errorf("write object magic: %w", err)
	}
	if _, err := writer.Write(noncePrefix); err != nil {
		return fmt.Errorf("write nonce prefix: %w", err)
	}

	buffer := make([]byte, chunkSize)
	for index := uint32(0); ; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		n, readErr := io.ReadFull(reader, buffer)
		final := false
		switch readErr {
		case nil:
		case io.EOF:
			final = true
			n = 0
		case io.ErrUnexpectedEOF:
			final = true
		default:
			return fmt.Errorf("read plaintext chunk: %w", readErr)
		}

		nonce := chunkNonce(noncePrefix, index)
		chunkAD := chunkAssociatedData(associatedData, index, final)
		ciphertext := aead.Seal(nil, nonce, buffer[:n], chunkAD)

		if err := writeChunkRecord(writer, final, uint32(n), ciphertext); err != nil {
			return err
		}
		if final {
			break
		}
	}
	return nil
}

func OpenObject(key []byte, encrypted []byte, associatedData []byte) ([]byte, error) {
	aead, err := fgcrypto.NewAES256GCM(key)
	if err != nil {
		return nil, err
	}
	if len(encrypted) < len(objectMagic)+noncePrefixSize {
		return nil, fmt.Errorf("encrypted object too small")
	}
	if string(encrypted[:len(objectMagic)]) != objectMagic {
		return nil, fmt.Errorf("invalid encrypted object magic")
	}

	reader := bytes.NewReader(encrypted[len(objectMagic):])
	noncePrefix := make([]byte, noncePrefixSize)
	if _, err := io.ReadFull(reader, noncePrefix); err != nil {
		return nil, fmt.Errorf("read nonce prefix: %w", err)
	}

	var plaintext bytes.Buffer
	for index := uint32(0); ; index++ {
		final, plainLen, ciphertext, err := readChunkRecord(reader, aead.Overhead())
		if err != nil {
			return nil, err
		}
		nonce := chunkNonce(noncePrefix, index)
		chunkAD := chunkAssociatedData(associatedData, index, final)
		chunk, err := aead.Open(nil, nonce, ciphertext, chunkAD)
		if err != nil {
			return nil, fmt.Errorf("open encrypted chunk %d: %w", index, err)
		}
		if len(chunk) != int(plainLen) {
			return nil, fmt.Errorf("chunk %d plaintext length mismatch", index)
		}
		plaintext.Write(chunk)
		if final {
			if reader.Len() != 0 {
				return nil, fmt.Errorf("trailing encrypted object data")
			}
			break
		}
	}
	return plaintext.Bytes(), nil
}

func writeChunkRecord(writer io.Writer, final bool, plaintextLen uint32, ciphertext []byte) error {
	var header [9]byte
	if final {
		header[0] = 1
	}
	binary.BigEndian.PutUint32(header[1:5], plaintextLen)
	binary.BigEndian.PutUint32(header[5:9], uint32(len(ciphertext)))
	if _, err := writer.Write(header[:]); err != nil {
		return fmt.Errorf("write chunk header: %w", err)
	}
	if _, err := writer.Write(ciphertext); err != nil {
		return fmt.Errorf("write chunk ciphertext: %w", err)
	}
	return nil
}

func readChunkRecord(reader *bytes.Reader, overhead int) (bool, uint32, []byte, error) {
	var header [9]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return false, 0, nil, fmt.Errorf("read chunk header: %w", err)
	}
	final := header[0] == 1
	plainLen := binary.BigEndian.Uint32(header[1:5])
	cipherLen := binary.BigEndian.Uint32(header[5:9])
	if cipherLen < uint32(overhead) {
		return false, 0, nil, fmt.Errorf("invalid chunk ciphertext length")
	}
	if cipherLen != plainLen+uint32(overhead) {
		return false, 0, nil, fmt.Errorf("chunk ciphertext length mismatch")
	}
	ciphertext := make([]byte, cipherLen)
	if _, err := io.ReadFull(reader, ciphertext); err != nil {
		return false, 0, nil, fmt.Errorf("read chunk ciphertext: %w", err)
	}
	return final, plainLen, ciphertext, nil
}

func chunkNonce(prefix []byte, index uint32) []byte {
	nonce := make([]byte, noncePrefixSize+4)
	copy(nonce, prefix)
	binary.BigEndian.PutUint32(nonce[noncePrefixSize:], index)
	return nonce
}

func chunkAssociatedData(base []byte, index uint32, final bool) []byte {
	var output bytes.Buffer
	output.Write(base)
	output.WriteByte(0)
	var indexBytes [4]byte
	binary.BigEndian.PutUint32(indexBytes[:], index)
	output.Write(indexBytes[:])
	if final {
		output.WriteByte(1)
	} else {
		output.WriteByte(0)
	}
	return output.Bytes()
}
