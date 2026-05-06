export namespace main {

	export class AppInfo {
	    productName: string;
	    appId: string;
	    nativeFormatVersion: string;
	    schemaVersion: number;
	    dataDir: string;
	    cliExecutableName: string;
	    cliShortAlias: string;

	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.productName = source["productName"];
	        this.appId = source["appId"];
	        this.nativeFormatVersion = source["nativeFormatVersion"];
	        this.schemaVersion = source["schemaVersion"];
	        this.dataDir = source["dataDir"];
	        this.cliExecutableName = source["cliExecutableName"];
	        this.cliShortAlias = source["cliShortAlias"];
	    }
	}
	export class CreateProjectRequest {
	    sourcePath: string;
	    contentOutput: string;
	    password: string;
	    maxPartSize: number;
	    force: boolean;
	    sourceCleanup: string;
	    databaseExport: string;

	    static createFrom(source: any = {}) {
	        return new CreateProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.contentOutput = source["contentOutput"];
	        this.password = source["password"];
	        this.maxPartSize = source["maxPartSize"];
	        this.force = source["force"];
	        this.sourceCleanup = source["sourceCleanup"];
	        this.databaseExport = source["databaseExport"];
	    }
	}
	export class CreateProjectResult {
	    projectId: string;
	    projectName: string;
	    contentOutput: string;
	    databaseExport: string;
	    encryptedFiles: number;
	    encryptedFolders: number;
	    encryptedParts: number;
	    deletedCleartextFiles: number;
	    deletedCleartextFolders: number;
	    failedFiles: number;

	    static createFrom(source: any = {}) {
	        return new CreateProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.contentOutput = source["contentOutput"];
	        this.databaseExport = source["databaseExport"];
	        this.encryptedFiles = source["encryptedFiles"];
	        this.encryptedFolders = source["encryptedFolders"];
	        this.encryptedParts = source["encryptedParts"];
	        this.deletedCleartextFiles = source["deletedCleartextFiles"];
	        this.deletedCleartextFolders = source["deletedCleartextFolders"];
	        this.failedFiles = source["failedFiles"];
	    }
	}
	export class DeleteProjectRequest {
	    projectId: string;
	    password: string;

	    static createFrom(source: any = {}) {
	        return new DeleteProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	    }
	}
	export class DeleteProjectResult {
	    projectId: string;

	    static createFrom(source: any = {}) {
	        return new DeleteProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	    }
	}
	export class ExportProjectRequest {
	    projectId: string;
	    password: string;
	    outputPath: string;
	    force: boolean;

	    static createFrom(source: any = {}) {
	        return new ExportProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	        this.outputPath = source["outputPath"];
	        this.force = source["force"];
	    }
	}
	export class ExportProjectResult {
	    projectId: string;
	    outputPath: string;

	    static createFrom(source: any = {}) {
	        return new ExportProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.outputPath = source["outputPath"];
	    }
	}
	export class ImportProjectRequest {
	    inputPath: string;
	    password: string;
	    force: boolean;

	    static createFrom(source: any = {}) {
	        return new ImportProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputPath = source["inputPath"];
	        this.password = source["password"];
	        this.force = source["force"];
	    }
	}
	export class ImportProjectResult {
	    projectId: string;

	    static createFrom(source: any = {}) {
	        return new ImportProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	    }
	}
	export class InspectProjectRequest {
	    projectId: string;
	    password: string;

	    static createFrom(source: any = {}) {
	        return new InspectProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	    }
	}
	export class InspectProjectResult {
	    projectId: string;
	    databaseType: string;
	    rootFolderId: string;
	    rootName: string;
	    formatVersion: string;
	    schemaVersion: string;
	    items: number;
	    folders: number;
	    files: number;
	    parts: number;
	    storageObjects: number;

	    static createFrom(source: any = {}) {
	        return new InspectProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.databaseType = source["databaseType"];
	        this.rootFolderId = source["rootFolderId"];
	        this.rootName = source["rootName"];
	        this.formatVersion = source["formatVersion"];
	        this.schemaVersion = source["schemaVersion"];
	        this.items = source["items"];
	        this.folders = source["folders"];
	        this.files = source["files"];
	        this.parts = source["parts"];
	        this.storageObjects = source["storageObjects"];
	    }
	}
	export class LocalProjectSummary {
	    projectId: string;
	    fileName: string;
	    modifiedAt: string;
	    availabilityStatus: string;

	    static createFrom(source: any = {}) {
	        return new LocalProjectSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.fileName = source["fileName"];
	        this.modifiedAt = source["modifiedAt"];
	        this.availabilityStatus = source["availabilityStatus"];
	    }
	}
	export class Settings {
	    operationGuideFormat: string;
	    defaultMaxPartSize: number;
	    sourceCleanupMode: string;
	    rememberRecentPaths: boolean;
	    recentPaths: string[];
	    windowStatePersistence: boolean;
	    theme: string;
	    language: string;

	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationGuideFormat = source["operationGuideFormat"];
	        this.defaultMaxPartSize = source["defaultMaxPartSize"];
	        this.sourceCleanupMode = source["sourceCleanupMode"];
	        this.rememberRecentPaths = source["rememberRecentPaths"];
	        this.recentPaths = source["recentPaths"];
	        this.windowStatePersistence = source["windowStatePersistence"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	    }
	}
	export class VerifyProjectRequest {
	    projectId: string;
	    password: string;
	    encryptedPath: string;

	    static createFrom(source: any = {}) {
	        return new VerifyProjectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	    }
	}
	export class VerifyProjectResult {
	    projectId: string;
	    checkedObjects: number;
	    missingObjects: number;
	    tamperedObjects: number;
	    extraObjects: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new VerifyProjectResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.checkedObjects = source["checkedObjects"];
	        this.missingObjects = source["missingObjects"];
	        this.tamperedObjects = source["tamperedObjects"];
	        this.extraObjects = source["extraObjects"];
	        this.status = source["status"];
	    }
	}

}
