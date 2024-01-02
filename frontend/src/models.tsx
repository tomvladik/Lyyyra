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
}

export interface Author {
    Type: string
    Value: string
}