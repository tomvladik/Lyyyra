
export interface AppStatus {
    WebResourcesReady: boolean;
    SongsReady: boolean;
    DatabaseReady: boolean;
    IsProgress: boolean;
    LastSave: string;
}
export const isEqualAppStatus = (status1: AppStatus, status2: AppStatus): boolean => {
    return (
        status1.WebResourcesReady === status2.WebResourcesReady &&
        status1.SongsReady === status2.SongsReady &&
        status1.IsProgress === status2.IsProgress &&
        status1.DatabaseReady === status2.DatabaseReady
    );
};