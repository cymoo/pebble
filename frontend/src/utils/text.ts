/**
 * Converts a text string to an HTML DOM element
 * Uses DOMParser to parse the text as HTML and returns the body element
 *
 * @param text - The text string to be converted to HTML
 * @returns The body element of the parsed HTML document
 */
export function textToHtml(text: string) {
  return new DOMParser().parseFromString(text, 'text/html').body
}

/**
 * Count English words and Chinese characters in a string
 * @param text Input string containing English and Chinese
 * @returns number Total count of English words and Chinese characters
 */
export function countWords(text: string): number {
  // Helper function to check if a character is Chinese
  const isChinese = (char: string): boolean => {
    const code = char.charCodeAt(0)
    return code >= 0x4e00 && code <= 0x9fff
  }

  // Remove HTML tags
  text = text.replace(/(<([^>]+)>)/gi, ' ')

  // Count Chinese characters
  const chineseCount = Array.from(text).filter((char) => isChinese(char)).length

  // Count English words (split by spaces and filter out empty strings)
  const englishWords = text
    .replace(/[^\w\s]/g, ' ') // Replace non-English chars with space
    .split(/\s+/) // Split by whitespace
    .filter((word) => word.length > 0).length // Remove empty strings

  return chineseCount + englishWords
}

/**
 * Checks if a string represents a valid integer
 */
export function isInteger(str: string): boolean {
  return /^-?\d+$/.test(str)
}
