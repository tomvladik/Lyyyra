
export interface SongData {
    songNumber: number;
    title: string;
    authorText: string;
    authorMusic: string;
    lyrics: string;
}

export interface dtoSong {
    Id: number
    Entry: number
    Title: string
    Verses: string
    AuthorMusic: string
    AuthorLyric: string
    KytaraFile: string
    SongbookAcronym: string
}

export interface SelectedSong {
    id: number;
    entry: number;
    title: string;
    filename?: string;
    hasNotes: boolean;
}

export interface dtoSongHeader {
    Id: number
    Entry: number
    Title: string
    TitleD: string
    KytaraFile: string
}

export interface Author {
    Type: string
    Value: string
}
export interface SelectParams {
    filter: string;
    orderBy: string;
}
