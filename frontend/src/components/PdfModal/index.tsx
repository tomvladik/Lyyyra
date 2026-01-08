import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { GetPdfFile } from "../../../wailsjs/go/app/App";
import styles from "./index.module.less";

interface PdfModalProps {
    isOpen: boolean;
    filename?: string;
    dataUrl?: string;
    songNumber?: number;
    songName?: string;
    onClose: () => void;
}

export const PdfModal = ({ isOpen, filename, dataUrl, songNumber, songName, onClose }: PdfModalProps) => {
    const { t } = useTranslation();
    const [pdfPath, setPdfPath] = useState<string>("");
    const [error, setError] = useState<string>("");
    const modalTitle = Number.isFinite(songNumber) ? `${songNumber} – ${songName}` : songName || filename || "PDF";

    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                onClose();
            }
        };

        if (isOpen) {
            document.addEventListener("keydown", handleEscape);
            return () => document.removeEventListener("keydown", handleEscape);
        }
    }, [isOpen, onClose]);

    useEffect(() => {
        if (!isOpen) {
            return;
        }

        setError("");
        setPdfPath("");

        if (dataUrl) {
            setPdfPath(dataUrl);
            return;
        }

        if (!filename) {
            setError("PDF source is missing");
            return;
        }

        GetPdfFile(filename)
            .then((remoteDataUrl: string) => {
                setPdfPath(remoteDataUrl);
            })
            .catch((err: Error) => {
                console.error("Failed to get PDF path:", err);
                setError(`Failed to load PDF: ${err}`);
            });
    }, [isOpen, filename, dataUrl]);

    if (!isOpen) return null;

    return (
        <div className={styles.modalOverlay} onClick={onClose}>
            <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
                <div className={styles.modalHeader}>
                    <h2 className={styles.modalTitle}>{modalTitle}</h2>
                    <button className={styles.closeButton} onClick={onClose} title={t('pdfModal.close')}>
                        ✕
                    </button>
                </div>
                <div className={styles.pdfContainer}>
                    {error ? (
                        <div className={styles.errorMessage}>{error}</div>
                    ) : pdfPath ? (
                        <iframe
                            src={pdfPath}
                            className={styles.pdfFrame}
                            title={`PDF: ${filename}`}
                        />
                    ) : (
                        <div className={styles.loadingMessage}>{t('pdfModal.loadingPdf')}</div>
                    )}
                </div>
            </div>
        </div>
    );
};
