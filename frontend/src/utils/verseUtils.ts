/**
 * Parses verse text into an array of formatted verse strings.
 * Splits on multiple newlines and normalizes whitespace.
 * 
 * @param versesText - Raw verse text from the database
 * @returns Array of cleaned verse strings
 */
export function parseVerses(versesText: string): string[] {
    if (!versesText) return [];

    return versesText
        .split(/\n\n+/)
        .map(v => v.replace(/\r?\n+/g, ' ').replace(/\s+/g, ' ').trim())
        .filter(Boolean);
}
