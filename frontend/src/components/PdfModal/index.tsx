import { useEffect, useState } from "react";
import styles from "./index.module.less";

interface WindowWithWails extends Window {
    go?: {
        main?: {
            App?: {
                GetPdfFile?: (filename: string) => Promise<string>;
            };
        };
    };
}

interface PdfModalProps {
    isOpen: boolean;
    filename: string;
    onClose: () => void;
}

const invokeGetPdfFile = (filename: string) => {
    const getPdfFile = (window as WindowWithWails).go?.main?.App?.GetPdfFile;

    if (!getPdfFile) {
        return Promise.reject(new Error("GetPdfFile backend method is unavailable"));
    }

    return getPdfFile(filename);
};

export const PdfModal = ({ isOpen, filename, onClose }: PdfModalProps) => {
    const [pdfPath, setPdfPath] = useState<string>("");
    const [error, setError] = useState<string>("");

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
        if (isOpen && filename) {
            setError("");
            setPdfPath("");
            invokeGetPdfFile(filename)
                .then((dataUrl: string) => {
                    setPdfPath(dataUrl);
                })
                .catch((err: any) => {
                    console.error("Failed to get PDF path:", err);
                    setError(`Failed to load PDF: ${err}`);
                });
        }
    }, [isOpen, filename]);

    if (!isOpen) return null;

    return (
        <div className={styles.modalOverlay} onClick={onClose}>
            <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
                <div className={styles.modalHeader}>
                    <h2 className={styles.modalTitle}>{filename}</h2>
                    <button className={styles.closeButton} onClick={onClose} title="Zavřít (Esc)">
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
                        <div className={styles.loadingMessage}>Načítání PDF...</div>
                    )}
                </div>
            </div>
        </div>
    );
};
