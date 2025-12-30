import { AppStatus, SortingOption } from '../AppStatus';

/**
 * Creates a mock AppStatus object for testing with optional overrides
 */
export const createMockStatus = (overrides: Partial<AppStatus> = {}): AppStatus => ({
    DatabaseReady: false,
    SongsReady: false,
    WebResourcesReady: false,
    IsProgress: false,
    ProgressMessage: '',
    ProgressPercent: 0,
    LastSave: '',
    SearchPattern: '',
    Sorting: 'entry' as SortingOption,
    ...overrides
});
