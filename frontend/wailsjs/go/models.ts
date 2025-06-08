export namespace main {
	
	export class AppStatus {
	    WebResourcesReady: boolean;
	    SongsReady: boolean;
	    DatabaseReady: boolean;
	    IsProgress: boolean;
	    // Go type: time
	    LastSave: any;
	    Sorting: string;
	    SearchPattern: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WebResourcesReady = source["WebResourcesReady"];
	        this.SongsReady = source["SongsReady"];
	        this.DatabaseReady = source["DatabaseReady"];
	        this.IsProgress = source["IsProgress"];
	        this.LastSave = this.convertValues(source["LastSave"], null);
	        this.Sorting = source["Sorting"];
	        this.SearchPattern = source["SearchPattern"];
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
	export class Author {
	    Type: string;
	    Value: string;
	
	    static createFrom(source: any = {}) {
	        return new Author(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Value = source["Value"];
	    }
	}
	export class dtoSong {
	    Id: number;
	    Entry: number;
	    Title: string;
	    Verses: string;
	    AuthorMusic: string;
	    AuthorLyric: string;
	
	    static createFrom(source: any = {}) {
	        return new dtoSong(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Entry = source["Entry"];
	        this.Title = source["Title"];
	        this.Verses = source["Verses"];
	        this.AuthorMusic = source["AuthorMusic"];
	        this.AuthorLyric = source["AuthorLyric"];
	    }
	}
	export class dtoSongHeader {
	    Id: number;
	    Entry: number;
	    Title: string;
	    TitleD: string;
	
	    static createFrom(source: any = {}) {
	        return new dtoSongHeader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Entry = source["Entry"];
	        this.Title = source["Title"];
	        this.TitleD = source["TitleD"];
	    }
	}

}

