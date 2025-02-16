
export interface AppStatus {
    SearchPattern: string;
    WebResourcesReady: boolean;
    SongsReady: boolean;
    DatabaseReady: boolean;
    IsProgress: boolean;
    LastSave: string;
    Sorting: SortingOption;
}

export type SortingOption = 'entry' | 'title' | 'authorMusic' | 'authorLyric';

export const isEqualAppStatus = (status1: AppStatus, status2: AppStatus): boolean => {
    return (
        status1.WebResourcesReady === status2.WebResourcesReady &&
        status1.SongsReady === status2.SongsReady &&
        status1.IsProgress === status2.IsProgress &&
        status1.DatabaseReady === status2.DatabaseReady &&
        status1.Sorting === status2.Sorting &&
        status1.SearchPattern === status2.SearchPattern
    );
};