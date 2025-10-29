// Cache storage for mapping URLs to File objects
// https://stackoverflow.com/questions/11876175/how-to-get-a-file-or-blob-from-an-object-url
const caches = new Map<string, File>()

/**
 * Utility for managing Object URLs with an associated File storage.
 * Extends the native URL.createObjectURL/revokeObjectURL functionality
 * by maintaining references to the original File objects.
 */
export const URLWithStore = {
  /**
   * Creates a blob URL for a File and stores the File reference.
   * Wrapper around URL.createObjectURL that maintains the File in cache.
   *
   * @param file - The File object to create a URL for
   * @returns A blob URL string pointing to the file content
   */
  createObjectURL(file: File): string {
    const url = URL.createObjectURL(file)
    caches.set(url, file)
    return url
  },

  /**
   * Revokes a blob URL and removes the associated File from cache.
   * Only removes from cache if the URL uses the blob: protocol.
   *
   * @param url - The URL string to revoke
   */
  revokeObjectURL(url: string): void {
    URL.revokeObjectURL(url)
    if (new URL(url).protocol === 'blob:') {
      caches.delete(url)
    }
  },

  /**
   * Retrieves the original File object associated with a URL
   *
   * @param url - The URL string to look up
   * @returns The associated File object, or null if not found
   */
  getFile(url: string): File | null {
    return caches.get(url) ?? null
  },
}
