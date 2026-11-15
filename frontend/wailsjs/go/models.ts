export namespace main {
	
	export class fileData {
	    name: string;
	    isFolder: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new fileData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isFolder = source["isFolder"];
	        this.size = source["size"];
	    }
	}

}

