export namespace app {
	
	export class AppStatus {
	    WebResourcesReady: boolean;
	    SongsReady: boolean;
	    DatabaseReady: boolean;
	    IsProgress: boolean;
	    ProgressMessage: string;
	    ProgressKey: string;
	    ProgressCurrent: number;
	    ProgressTotal: number;
	    ProgressPercent: number;
	    LastSave: string;
	    Sorting: string;
	    SearchPattern: string;
	    BuildVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WebResourcesReady = source["WebResourcesReady"];
	        this.SongsReady = source["SongsReady"];
	        this.DatabaseReady = source["DatabaseReady"];
	        this.IsProgress = source["IsProgress"];
	        this.ProgressMessage = source["ProgressMessage"];
	        this.ProgressKey = source["ProgressKey"];
	        this.ProgressCurrent = source["ProgressCurrent"];
	        this.ProgressTotal = source["ProgressTotal"];
	        this.ProgressPercent = source["ProgressPercent"];
	        this.LastSave = source["LastSave"];
	        this.Sorting = source["Sorting"];
	        this.SearchPattern = source["SearchPattern"];
	        this.BuildVersion = source["BuildVersion"];
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
	    KytaraFile: string;
	    VarhanyFile: string;
	    SongbookAcronym: string;
	
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
	        this.KytaraFile = source["KytaraFile"];
	        this.VarhanyFile = source["VarhanyFile"];
	        this.SongbookAcronym = source["SongbookAcronym"];
	    }
	}
	export class dtoSongHeader {
	    Id: number;
	    Entry: number;
	    Title: string;
	    TitleD: string;
	    KytaraFile: string;
	    VarhanyFile: string;
	
	    static createFrom(source: any = {}) {
	        return new dtoSongHeader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Entry = source["Entry"];
	        this.Title = source["Title"];
	        this.TitleD = source["TitleD"];
	        this.KytaraFile = source["KytaraFile"];
	        this.VarhanyFile = source["VarhanyFile"];
	    }
	}

}

