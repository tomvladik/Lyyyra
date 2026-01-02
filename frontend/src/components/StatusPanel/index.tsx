import React, { useContext, useState } from "react";
import styles from "./index.module.less";

import { GetStatus, ResetData } from "../../../wailsjs/go/main/App";
import { AppStatus, SortingOption } from "../../AppStatus";
import { DataContext } from "../../context";

interface StatusPanelProps {
    onHide?: () => void;
}

const StatusPanel: React.FC<StatusPanelProps> = ({ onHide }) => {
    const { status: data, updateStatus } = useContext(DataContext);
    const [isResetting, setIsResetting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const formatDate = (value: string) => {
        if (!value) {
            return "---";
        }
        const parsed = new Date(value);
        if (Number.isNaN(parsed.getTime())) {
            return value;
        }
        return parsed.toLocaleString("cs-CZ", {
            hour: "2-digit",
            minute: "2-digit",
            day: "2-digit",
            month: "2-digit",
        });
    };

    const progressPercent = Math.max(0, Math.min(100, data.ProgressPercent || 0));
    const isBusy = data.IsProgress;
    const progressMessage = data.ProgressMessage?.trim() || (isBusy ? "Pracuji..." : "Hotovo");
    const versionLabel = data.BuildVersion || "dev";
    const lastSaveLabel = formatDate(data.LastSave);
    const allReady = data.WebResourcesReady && data.DatabaseReady && data.SongsReady;
    const showProgress = isBusy || !allReady;

    const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
        onHide?.();
    };

    const handleDoubleClick = (event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
    };

    const handleReset = async () => {
        if (isResetting) return;

        const confirmed = window.confirm("Smazat všechna stažená data a znovu inicializovat?");
        if (!confirmed) return;

        setIsResetting(true);
        setError(null);
        updateStatus({
            IsProgress: true,
            ProgressMessage: "Mažu uložená data...",
            ProgressPercent: 0,
            WebResourcesReady: false,
            DatabaseReady: false,
            SongsReady: false,
        });

        try {
            await ResetData();
            const finalStatus = await GetStatus();
            updateStatus({
                ...finalStatus,
                Sorting: (finalStatus.Sorting || "entry") as SortingOption,
                ProgressMessage: finalStatus.ProgressMessage || "",
                ProgressPercent: finalStatus.ProgressPercent || 0,
            } as Partial<AppStatus>);
        } catch (err) {
            console.error(err);
            setError("Nepodařilo se smazat data. Zkuste to prosím znovu.");
        } finally {
            setIsResetting(false);
        }
    };

    return (
        <div
            className={styles.statusPanel}
            onClick={handleClick}
            onDoubleClick={handleDoubleClick}
        >
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Verze:</span>
                <span className={styles.statusValueTight} title={versionLabel}>
                    {versionLabel}
                </span>
            </div>

            {showProgress && (
                <div className={styles.progressRow}>
                    <span className={styles.progressMessage} title={progressMessage}>{progressMessage}</span>
                    <span className={styles.statusValue}>{progressPercent}%</span>
                </div>
            )}
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Uloženo:</span>
                <span className={styles.statusValueTight} title={lastSaveLabel}>
                    {lastSaveLabel}
                </span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Podklady:</span>
                <span className={styles.statusValue}>{data.WebResourcesReady ? "OK" : "čekám"}</span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Databáze:</span>
                <span className={styles.statusValue}>{data.DatabaseReady ? "OK" : "čekám"}</span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Skladby:</span>
                <span className={styles.statusValue}>{data.SongsReady ? "OK" : "čekám"}</span>
            </div>
            <div className={styles.statusRow} >

                <button
                    className={styles.resetButton}
                    onClick={handleReset}
                    disabled={isResetting || isBusy}
                    data-testid="reset-data-button"
                    title="Smazat data a znovu inicializovat"
                    aria-label="Smazat data a znovu inicializovat"
                >
                    <svg
                        width="16"
                        height="16"
                        viewBox="0 0 16 16"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                        aria-hidden="true"
                        focusable="false"
                        className={styles.resetIcon}
                    >
                        <path
                            d="M4 4.5C4.8 3.6 5.9 3 7.1 3c2.2 0 4 1.8 4 4 0 2.2-1.8 4-4 4-1.3 0-2.5-.6-3.2-1.6"
                            stroke="currentColor"
                            strokeWidth="1.2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                        />
                        <path
                            d="M4.5 2.5v2h-2"
                            stroke="currentColor"
                            strokeWidth="1.2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                        />
                    </svg>
                </button>
            </div>
            {error && <div className={styles.statusValue}>{error}</div>}
        </div>
    );
};

export default StatusPanel;
