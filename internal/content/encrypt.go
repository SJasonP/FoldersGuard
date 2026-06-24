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
	"strings"

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
	// OnBytes, when set, is called with the number of plaintext bytes processed
	// after each chunk is encrypted. It enables byte-weighted progress.
	OnBytes func(int64)
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
	if err := e.sealReader(ctx, aead, input, outputPath, associatedData); err != nil {
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
		if err := e.sealReader(ctx, aead, section, outputPath, associatedData); err != nil {
			return fmt.Errorf("encrypt part %d: %w", part.Index, err)
		}
	}
	return nil
}

func (e Encryptor) sealReader(ctx context.Context, aead cipher.AEAD, reader io.Reader, outputPath string, associatedData []byte) error {
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

	if err := writeEncryptedObject(ctx, aead, reader, output, associatedData, e.ChunkSize, e.OnBytes); err != nil {
		return err
	}
	return nil
}

func writeEncryptedObject(ctx context.Context, aead cipher.AEAD, reader io.Reader, writer io.Writer, associatedData []byte, chunkSize int, onProgress func(int64)) error {
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
		if n > 0 && onProgress != nil {
			onProgress(int64(n))
		}
		if final {
			break
		}
	}
	return nil
}

// StreamDecrypt reads one encrypted object from reader, authenticates and
// decrypts it chunk by chunk, and writes the plaintext to writer without
// buffering the whole object in memory. When onProgress is set it is called with
// the number of plaintext bytes after each chunk.
func StreamDecrypt(ctx context.Context, key []byte, reader io.Reader, writer io.Writer, associatedData []byte, onProgress func(int64)) error {
	aead, err := fgcrypto.NewAES256GCM(key)
	if err != nil {
		return err
	}

	magic := make([]byte, len(objectMagic))
	if _, err := io.ReadFull(reader, magic); err != nil {
		return fmt.Errorf("read object magic: %w", err)
	}
	if string(magic) != objectMagic {
		return fmt.Errorf("invalid encrypted object magic")
	}

	noncePrefix := make([]byte, noncePrefixSize)
	if _, err := io.ReadFull(reader, noncePrefix); err != nil {
		return fmt.Errorf("read nonce prefix: %w", err)
	}

	for index := uint32(0); ; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		final, plainLen, ciphertext, err := readChunkRecordReader(reader, aead.Overhead())
		if err != nil {
			return err
		}
		nonce := chunkNonce(noncePrefix, index)
		chunkAD := chunkAssociatedData(associatedData, index, final)
		chunk, err := aead.Open(nil, nonce, ciphertext, chunkAD)
		if err != nil {
			return fmt.Errorf("open encrypted chunk %d: %w", index, err)
		}
		if len(chunk) != int(plainLen) {
			return fmt.Errorf("chunk %d plaintext length mismatch", index)
		}
		if len(chunk) > 0 {
			if _, err := writer.Write(chunk); err != nil {
				return fmt.Errorf("write plaintext chunk %d: %w", index, err)
			}
			if onProgress != nil {
				onProgress(int64(len(chunk)))
			}
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

func OpenObjectFile(ctx context.Context, key []byte, encryptedPath, outputPath string, associatedData []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if encryptedPath == "" {
		return fmt.Errorf("encrypted path is required")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	encrypted, err := os.ReadFile(encryptedPath)
	if err != nil {
		return fmt.Errorf("read encrypted object: %w", err)
	}
	plaintext, err := OpenObject(key, encrypted, associatedData)
	if err != nil {
		return err
	}
	if err := WritePlaintextFile(outputPath, plaintext); err != nil {
		return err
	}
	return nil
}

func OpenObjectFromFile(ctx context.Context, key []byte, encryptedPath string, associatedData []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if encryptedPath == "" {
		return nil, fmt.Errorf("encrypted path is required")
	}
	encrypted, err := os.ReadFile(encryptedPath)
	if err != nil {
		return nil, fmt.Errorf("read encrypted object: %w", err)
	}
	plaintext, err := OpenObject(key, encrypted, associatedData)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// OpenObjectFileStream decrypts the encrypted object at encryptedPath and writes
// the plaintext to outputPath, streaming chunk by chunk so the whole object is
// never held in memory. The output is written to a temporary file and committed
// atomically. When onProgress is set it reports plaintext bytes as they decrypt.
func OpenObjectFileStream(ctx context.Context, key []byte, encryptedPath, outputPath string, associatedData []byte, onProgress func(int64)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if encryptedPath == "" {
		return fmt.Errorf("encrypted path is required")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}
	input, err := os.Open(encryptedPath)
	if err != nil {
		return fmt.Errorf("open encrypted object: %w", err)
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create plaintext directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary plaintext: %w", err)
	}
	tempPath := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	if err := StreamDecrypt(ctx, key, input, temp, associatedData, onProgress); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restrict temporary plaintext permissions: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary plaintext: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("commit plaintext: %w", err)
	}
	committed = true
	return nil
}

// VerifyObjectFileStream authenticates and decrypts the encrypted object at
// encryptedPath without writing the plaintext, streaming chunk by chunk. It
// returns an error if the object cannot be authenticated. When onProgress is set
// it reports plaintext bytes as they are verified.
func VerifyObjectFileStream(ctx context.Context, key []byte, encryptedPath string, associatedData []byte, onProgress func(int64)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if encryptedPath == "" {
		return fmt.Errorf("encrypted path is required")
	}
	input, err := os.Open(encryptedPath)
	if err != nil {
		return fmt.Errorf("open encrypted object: %w", err)
	}
	defer input.Close()
	return StreamDecrypt(ctx, key, input, io.Discard, associatedData, onProgress)
}

func WritePlaintextFile(outputPath string, plaintext []byte) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create plaintext directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary plaintext: %w", err)
	}
	tempPath := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := temp.Write(plaintext); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary plaintext: %w", err)
	}
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restrict temporary plaintext permissions: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary plaintext: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("commit plaintext: %w", err)
	}
	committed = true
	return nil
}

func SafeJoin(root, relative string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("root path is required")
	}
	if relative == "" {
		return "", fmt.Errorf("relative path is required")
	}
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("absolute relative path rejected")
	}
	cleanRelative := filepath.Clean(filepath.FromSlash(relative))
	if cleanRelative == "." || cleanRelative == ".." || strings.HasPrefix(cleanRelative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return filepath.Join(root, cleanRelative), nil
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

func readChunkRecordReader(reader io.Reader, overhead int) (bool, uint32, []byte, error) {
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
