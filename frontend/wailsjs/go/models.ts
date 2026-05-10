export namespace main {
	
	export class AppInfo {
	    productName: string;
	    productVersion: string;
	    appId: string;
	    formatVersion: string;
	    dataDir: string;
	    startupError: string;
	    copyrightNotice: string;
	    projectLink: string;
	    thirdPartyLink: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.productName = source["productName"];
	        this.productVersion = source["productVersion"];
	        this.appId = source["appId"];
	        this.formatVersion = source["formatVersion"];
	        this.dataDir = source["dataDir"];
	        this.startupError = source["startupError"];
	        this.copyrightNotice = source["copyrightNotice"];
	        this.projectLink = source["projectLink"];
	        this.thirdPartyLink = source["thirdPartyLink"];
	    }
	}
	export class ProjectCreateFolderChange {
	    targetFolderPath: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectCreateFolderChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.targetFolderPath = source["targetFolderPath"];
	        this.name = source["name"];
	    }
	}
	export class ProjectAddChange {
	    sourcePath: string;
	    targetFolderPath: string;
	    maxPartSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectAddChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.targetFolderPath = source["targetFolderPath"];
	        this.maxPartSize = source["maxPartSize"];
	    }
	}
	export class ProjectRemoveChange {
	    itemPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectRemoveChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemPath = source["itemPath"];
	    }
	}
	export class ProjectMoveChange {
	    itemPath: string;
	    targetFolderPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectMoveChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemPath = source["itemPath"];
	        this.targetFolderPath = source["targetFolderPath"];
	    }
	}
	export class ProjectRenameChange {
	    itemPath: string;
	    newName: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectRenameChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemPath = source["itemPath"];
	        this.newName = source["newName"];
	    }
	}
	export class ApplyProjectChangesRequest {
	    projectId: string;
	    password: string;
	    encryptedPath: string;
	    renameChanges: ProjectRenameChange[];
	    moveChanges: ProjectMoveChange[];
	    removeChanges: ProjectRemoveChange[];
	    addChanges: ProjectAddChange[];
	    createFolderChanges: ProjectCreateFolderChange[];
	
	    static createFrom(source: any = {}) {
	        return new ApplyProjectChangesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	        this.renameChanges = this.convertValues(source["renameChanges"], ProjectRenameChange);
	        this.moveChanges = this.convertValues(source["moveChanges"], ProjectMoveChange);
	        this.removeChanges = this.convertValues(source["removeChanges"], ProjectRemoveChange);
	        this.addChanges = this.convertValues(source["addChanges"], ProjectAddChange);
	        this.createFolderChanges = this.convertValues(source["createFolderChanges"], ProjectCreateFolderChange);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectBrowserItem {
	    id: string;
	    parentId: string;
	    path: string;
	    parentPath: string;
	    name: string;
	    type: string;
	    size: number;
	    childCount: number;
	    modifiedAt: string;
	    metadataCaptured: boolean;
	    contentAvailable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectBrowserItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parentId = source["parentId"];
	        this.path = source["path"];
	        this.parentPath = source["parentPath"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.childCount = source["childCount"];
	        this.modifiedAt = source["modifiedAt"];
	        this.metadataCaptured = source["metadataCaptured"];
	        this.contentAvailable = source["contentAvailable"];
	    }
	}
	export class ProjectBrowserState {
	    projectId: string;
	    projectName: string;
	    rootFolderId: string;
	    rootFolderName: string;
	    createdAt: string;
	    updatedAt: string;
	    files: number;
	    folders: number;
	    parts: number;
	    contentConnected: boolean;
	    encryptedPath: string;
	    items: ProjectBrowserItem[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectBrowserState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.rootFolderId = source["rootFolderId"];
	        this.rootFolderName = source["rootFolderName"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.files = source["files"];
	        this.folders = source["folders"];
	        this.parts = source["parts"];
	        this.contentConnected = source["contentConnected"];
	        this.encryptedPath = source["encryptedPath"];
	        this.items = this.convertValues(source["items"], ProjectBrowserItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectContentOperation {
	    type: string;
	    sourcePath: string;
	    targetPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectContentOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.sourcePath = source["sourcePath"];
	        this.targetPath = source["targetPath"];
	    }
	}
	export class ApplyProjectChangesResult {
	    projectId: string;
	    appliedRenames: number;
	    appliedMoves: number;
	    appliedRemoves: number;
	    appliedAdds: number;
	    appliedCreatedFolders: number;
	    operationGuidePath: string;
	    stagedContentPath: string;
	    contentOperations: ProjectContentOperation[];
	    appliedContentChanges: ProjectContentOperation[];
	    browserState: ProjectBrowserState;
	
	    static createFrom(source: any = {}) {
	        return new ApplyProjectChangesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.appliedRenames = source["appliedRenames"];
	        this.appliedMoves = source["appliedMoves"];
	        this.appliedRemoves = source["appliedRemoves"];
	        this.appliedAdds = source["appliedAdds"];
	        this.appliedCreatedFolders = source["appliedCreatedFolders"];
	        this.operationGuidePath = source["operationGuidePath"];
	        this.stagedContentPath = source["stagedContentPath"];
	        this.contentOperations = this.convertValues(source["contentOperations"], ProjectContentOperation);
	        this.appliedContentChanges = this.convertValues(source["appliedContentChanges"], ProjectContentOperation);
	        this.browserState = this.convertValues(source["browserState"], ProjectBrowserState);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
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
	export class CreateShareRequest {
	    projectId: string;
	    projectPassword: string;
	    itemPaths: string[];
	    outputPath: string;
	    force: boolean;
	    passwordProtected: boolean;
	    sharePassword: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateShareRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectPassword = source["projectPassword"];
	        this.itemPaths = source["itemPaths"];
	        this.outputPath = source["outputPath"];
	        this.force = source["force"];
	        this.passwordProtected = source["passwordProtected"];
	        this.sharePassword = source["sharePassword"];
	    }
	}
	export class ShareContentLocation {
	    sourcePath: string;
	    targetPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ShareContentLocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.targetPath = source["targetPath"];
	    }
	}
	export class CreateShareResult {
	    projectId: string;
	    shareId: string;
	    outputPath: string;
	    topLevelItems: number;
	    files: number;
	    folders: number;
	    parts: number;
	    passwordProtected: boolean;
	    contentLocations: ShareContentLocation[];
	
	    static createFrom(source: any = {}) {
	        return new CreateShareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.shareId = source["shareId"];
	        this.outputPath = source["outputPath"];
	        this.topLevelItems = source["topLevelItems"];
	        this.files = source["files"];
	        this.folders = source["folders"];
	        this.parts = source["parts"];
	        this.passwordProtected = source["passwordProtected"];
	        this.contentLocations = this.convertValues(source["contentLocations"], ShareContentLocation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DecryptProjectRequest {
	    projectId: string;
	    password: string;
	    encryptedPath: string;
	    outputPath: string;
	    force: boolean;
	    sourceCleanup: string;
	
	    static createFrom(source: any = {}) {
	        return new DecryptProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	        this.outputPath = source["outputPath"];
	        this.force = source["force"];
	        this.sourceCleanup = source["sourceCleanup"];
	    }
	}
	export class DecryptProjectResult {
	    projectId: string;
	    outputPath: string;
	    decryptedFiles: number;
	    restoredFolders: number;
	    skippedFolders: number;
	    deletedEncryptedFiles: number;
	    failedEncryptedFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new DecryptProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.outputPath = source["outputPath"];
	        this.decryptedFiles = source["decryptedFiles"];
	        this.restoredFolders = source["restoredFolders"];
	        this.skippedFolders = source["skippedFolders"];
	        this.deletedEncryptedFiles = source["deletedEncryptedFiles"];
	        this.failedEncryptedFiles = source["failedEncryptedFiles"];
	    }
	}
	export class DecryptShareRequest {
	    databasePath: string;
	    password: string;
	    encryptedPath: string;
	    outputPath: string;
	    force: boolean;
	    sourceCleanup: string;
	
	    static createFrom(source: any = {}) {
	        return new DecryptShareRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.databasePath = source["databasePath"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	        this.outputPath = source["outputPath"];
	        this.force = source["force"];
	        this.sourceCleanup = source["sourceCleanup"];
	    }
	}
	export class DecryptShareResult {
	    shareId: string;
	    outputPath: string;
	    decryptedFiles: number;
	    restoredFolders: number;
	    skippedFolders: number;
	    deletedEncryptedFiles: number;
	    failedEncryptedFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new DecryptShareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shareId = source["shareId"];
	        this.outputPath = source["outputPath"];
	        this.decryptedFiles = source["decryptedFiles"];
	        this.restoredFolders = source["restoredFolders"];
	        this.skippedFolders = source["skippedFolders"];
	        this.deletedEncryptedFiles = source["deletedEncryptedFiles"];
	        this.failedEncryptedFiles = source["failedEncryptedFiles"];
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
	    projectName: string;
	    rootFolderId: string;
	    rootName: string;
	    formatVersion: string;
	    createdAt: string;
	    updatedAt: string;
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
	        this.projectName = source["projectName"];
	        this.rootFolderId = source["rootFolderId"];
	        this.rootName = source["rootName"];
	        this.formatVersion = source["formatVersion"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.items = source["items"];
	        this.folders = source["folders"];
	        this.files = source["files"];
	        this.parts = source["parts"];
	        this.storageObjects = source["storageObjects"];
	    }
	}
	export class LoadShareRequest {
	    databasePath: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoadShareRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.databasePath = source["databasePath"];
	        this.password = source["password"];
	    }
	}
	export class LocalProjectSummary {
	    projectId: string;
	    projectName: string;
	    fileName: string;
	    modifiedAt: string;
	    availabilityStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalProjectSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	        this.fileName = source["fileName"];
	        this.modifiedAt = source["modifiedAt"];
	        this.availabilityStatus = source["availabilityStatus"];
	    }
	}
	export class OpenProjectBrowserRequest {
	    projectId: string;
	    password: string;
	    encryptedPath: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenProjectBrowserRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	    }
	}
	export class PathDialogFilter {
	    displayName: string;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new PathDialogFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.pattern = source["pattern"];
	    }
	}
	
	
	
	
	
	
	
	
	export class SaveLocalProjectNameRequest {
	    projectId: string;
	    projectName: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveLocalProjectNameRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	    }
	}
	export class SaveLocalProjectNameResult {
	    projectId: string;
	    projectName: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveLocalProjectNameResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.projectName = source["projectName"];
	    }
	}
	export class SelectPathRequest {
	    kind: string;
	    title: string;
	    defaultDirectory: string;
	    defaultFilename: string;
	    filters: PathDialogFilter[];
	
	    static createFrom(source: any = {}) {
	        return new SelectPathRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.defaultDirectory = source["defaultDirectory"];
	        this.defaultFilename = source["defaultFilename"];
	        this.filters = this.convertValues(source["filters"], PathDialogFilter);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Settings {
	    operationGuideFormat: string;
	    defaultMaxPartSize: number;
	    sourceCleanupMode: string;
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
	        this.theme = source["theme"];
	        this.language = source["language"];
	    }
	}
	
	export class ShareSummary {
	    shareId: string;
	    databaseType: string;
	    formatVersion: string;
	    topLevelItems: number;
	    files: number;
	    folders: number;
	    parts: number;
	    storageObjects: number;
	    passwordProtected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ShareSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shareId = source["shareId"];
	        this.databaseType = source["databaseType"];
	        this.formatVersion = source["formatVersion"];
	        this.topLevelItems = source["topLevelItems"];
	        this.files = source["files"];
	        this.folders = source["folders"];
	        this.parts = source["parts"];
	        this.storageObjects = source["storageObjects"];
	        this.passwordProtected = source["passwordProtected"];
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
	export class VerifyShareRequest {
	    databasePath: string;
	    password: string;
	    encryptedPath: string;
	
	    static createFrom(source: any = {}) {
	        return new VerifyShareRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.databasePath = source["databasePath"];
	        this.password = source["password"];
	        this.encryptedPath = source["encryptedPath"];
	    }
	}

}

