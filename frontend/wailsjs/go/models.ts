export namespace main {
	
	export class PreviewPayload {
	    filePath: string;
	    html: string;
	    version: string;
	    renderedAt: string;
	    error?: string;
	    frontmatter?: any;
	
	    static createFrom(source: any = {}) {
	        return new PreviewPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.html = source["html"];
	        this.version = source["version"];
	        this.renderedAt = source["renderedAt"];
	        this.error = source["error"];
	        this.frontmatter = source["frontmatter"];
	    }
	}
	export class UpdateSettings {
	    autoUpdateEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoUpdateEnabled = source["autoUpdateEnabled"];
	    }
	}
	export class UpdateStatus {
	    state: string;
	    currentVersion: string;
	    latestVersion: string;
	    message: string;
	    downloadedPath: string;
	    releaseURL: string;
	    checkedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.message = source["message"];
	        this.downloadedPath = source["downloadedPath"];
	        this.releaseURL = source["releaseURL"];
	        this.checkedAt = source["checkedAt"];
	    }
	}

}

