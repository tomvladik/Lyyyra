import React, { useContext } from "react";
import styles from "./index.module.less";

import { DataContext } from "../../main";

interface StatusPanelProps {
    onHide?: () => void;
}

const StatusPanel: React.FC<StatusPanelProps> = ({ onHide }) => {
    const { status: data } = useContext(DataContext);

    const readiness = [
        { label: "Notové podklady", ready: data.WebResourcesReady },
        { label: "Databáze", ready: data.DatabaseReady },
        { label: "Skladby", ready: data.SongsReady },
    ];

    const formatDate = (value: string) => {
        if (!value) {
            return "---";
        }
        const parsed = new Date(value);
        if (Number.isNaN(parsed.getTime())) {
            return value;
        }
        return parsed.toLocaleString("cs-CZ", {
            weekday: "short",
            hour: "2-digit",
            minute: "2-digit",
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
        });
    };

    const progressPercent = Math.max(0, Math.min(100, data.ProgressPercent || 0));
    const isBusy = data.IsProgress && progressPercent < 100;
    const progressMessage = data.ProgressMessage?.trim() || (isBusy ? "Pracuji..." : "Hotovo");
    const versionLabel = data.BuildVersion || "dev";

    const handleHideClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        onHide?.();
    };

    const handleDoubleClick = (event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
    };

    return (
        <div className={styles.statusPanel} onDoubleClick={handleDoubleClick}>
            <div className={styles.panelHeader}>
                <div>
                    <p className={styles.panelEyebrow}>Diagnostika</p>
                    <h2 className={styles.panelTitle}>Stav aplikace</h2>
                </div>
                <button type="button" className={styles.hideButton} onClick={handleHideClick}>
                    Skrýt panel
                </button>
            </div>

            <div className={styles.readyGrid}>
                {readiness.map(({ label, ready }) => (
                    <div
                        key={label}
                        className={ready ? styles.statusChipReady : styles.statusChipPending}
                    >
                        <span className={styles.statusDot} />
                        <div>
                            <p className={styles.chipLabel}>{label}</p>
                            <p className={styles.chipState}>{ready ? "připraveno" : "čekám"}</p>
                        </div>
                    </div>
                ))}
            </div>

            <div className={styles.metaRow}>
                <div className={styles.metaGroup}>
                    <p className={styles.metaLabel}>Poslední uložení</p>
                    <p className={styles.metaValue}>{formatDate(data.LastSave)}</p>
                </div>
                <div className={styles.metaGroup}>
                    <p className={styles.metaLabel}>Verze</p>
                    <p className={styles.metaValue}>{versionLabel}</p>
                </div>
            </div>

            <div className={styles.progressBlock}>
                <div className={styles.progressHeader}>
                    <p className={styles.progressLabel}>{progressMessage}</p>
                    <p className={styles.progressValue}>{progressPercent}%</p>
                </div>
                <div className={styles.progressTrack}>
                    <div
                        className={styles.progressBar}
                        style={{ width: `${progressPercent}%` }}
                    />
                </div>
            </div>
        </div>
    );
};

export default StatusPanel;
