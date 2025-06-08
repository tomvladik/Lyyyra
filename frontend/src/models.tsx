
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
}

export interface dtoSongHeader {
    Id: number
    Entry: number
    Title: string
    TitleD: string
}

export interface Author {
    Type: string
    Value: string
}
export interface SelectParams {
    filter: string;
    orderBy: string;
}
