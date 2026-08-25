export namespace main {
	
	export class ProcessOptions {
	    mediaType: string;
	    mode: string;
	    value: string;
	    outputDir: string;
	    vramMode: string;
	    engine: string;
	    resolution: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaType = source["mediaType"];
	        this.mode = source["mode"];
	        this.value = source["value"];
	        this.outputDir = source["outputDir"];
	        this.vramMode = source["vramMode"];
	        this.engine = source["engine"];
	        this.resolution = source["resolution"];
	    }
	}

}

