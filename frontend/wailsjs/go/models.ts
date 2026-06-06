export namespace main {
	
	export class PreviewPayload {
	    filePath: string;
	    html: string;
	    version: string;
	    renderedAt: string;
	    error?: string;
	
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
	    }
	}

}

