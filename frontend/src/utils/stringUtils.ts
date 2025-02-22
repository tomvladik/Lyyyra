export function removeDiacritics(text: string): string {
    if (!text) return "";
    return text
        .normalize("NFKD") // Decomposes characters into base + diacritic
        .replace(/\p{M}/gu, ""); // Removes diacritic marks
}