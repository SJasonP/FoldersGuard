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

}

