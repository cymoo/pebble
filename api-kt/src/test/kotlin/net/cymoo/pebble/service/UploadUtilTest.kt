package net.cymoo.pebble.service

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Disabled
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class GenerateSecureFilenameTest {
    @Test
    fun `test generateSecureFilename with default length`() {
        val result = generateSecureFilename("test.txt")
        assertTrue(result.matches(Regex("test\\.[a-f0-9]{8}\\.txt")))
    }

    @Test
    fun `test generateSecureFilename with minimum length`() {
        val result = generateSecureFilename("test.txt", 8)
        assertTrue(result.matches(Regex("test\\.[a-f0-9]{8}\\.txt")))
    }

    @Test
    fun `test generateSecureFilename with custom length`() {
        val result = generateSecureFilename("test.txt", 16)
        assertTrue(result.matches(Regex("test\\.[a-f0-9]{16}\\.txt")))
    }

    @Test
    fun `test generateSecureFilename uniqueness with same length`() {
        val length = 12
        val results = List(100) {
            generateSecureFilename("test.txt", length)
        }

        val uniqueResults = results.toSet()
        assertEquals(100, uniqueResults.size, "Generated UUIDs should be unique")
    }

    @Test
    fun `test generateSecureFilename with Chinese characters`() {
        val result = generateSecureFilename("测试文件.pdf")
        assertTrue(result.matches(Regex("测试文件\\.[a-f0-9]{8}\\.pdf")))
    }

    @Test
    fun `test generateSecureFilename with special characters`() {
        val result = generateSecureFilename("test@#$%^&*.txt")
        assertTrue(result.matches(Regex("test_\\.[a-f0-9]{8}\\.txt")))
        assertFalse(result.contains(Regex("[^\\w\\-.\\u4e00-\\u9fa5]")))
    }

    @Test
    fun `test generateSecureFilename with no extension`() {
        val result = generateSecureFilename("testfile")
        assertTrue(result.matches(Regex("testfile\\.[a-f0-9]{8}")))
    }

    @Test
    fun `test generateSecureFilename with hidden file`() {
        val result = generateSecureFilename(".gitignore")
        assertTrue(result.matches(Regex("\\.gitignore\\.[a-f0-9]{8}")))
    }

    @Test
    fun `test generateSecureFilename with hidden file and extension`() {
        val result = generateSecureFilename(".config.json")
        assertTrue(result.matches(Regex("\\.config\\.[a-f0-9]{8}\\.json")))
    }

    @Test
    fun `test generateSecureFilename with multiple dots`() {
        val result = generateSecureFilename("test.backup.txt")
        assertTrue(result.matches(Regex("test\\.backup\\.[a-f0-9]{8}\\.txt")))
    }

    @Test
    @Disabled
    fun `test generateSecureFilename with common archive extensions`() {
        val extensions = listOf("tar.gz", "tar.bz2")
        for (ext in extensions) {
            val result = generateSecureFilename("archive.$ext")
            assertTrue(result.endsWith(ext), "Should preserve complex extension: $ext")
        }
    }

    @Test
    fun `test generateSecureFilename with very long filename`() {
        val longName = "a".repeat(200)
        val result = generateSecureFilename("$longName.txt")
        assertTrue(result.matches(Regex("a{200}\\.[a-f0-9]{8}\\.txt")))
    }

    @Test
    fun `test generateSecureFilename with invalid length throws exception`() {
        assertThrows<IllegalArgumentException> {
            generateSecureFilename("test.txt", 7)
        }
        assertThrows<IllegalArgumentException> {
            generateSecureFilename("test.txt", 33)
        }
    }

    @Test
    fun `test generateSecureFilename with blank filename throws exception`() {
        assertThrows<IllegalArgumentException> {
            generateSecureFilename("  ")
        }
    }

    @Test
    fun `test generateSecureFilename preserves case`() {
        val result = generateSecureFilename("TestFile.TXT")
        assertTrue(result.startsWith("TestFile."))
        assertTrue(result.endsWith(".TXT"))
    }

    @Test
    fun `test generateSecureFilename with empty extension`() {
        val result = generateSecureFilename("test.")
        assertTrue(result.matches(Regex("test\\.[a-f0-9]{8}")))
    }
}
