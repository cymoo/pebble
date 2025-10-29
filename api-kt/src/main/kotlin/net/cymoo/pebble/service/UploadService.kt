package net.cymoo.pebble.service

import com.drew.imaging.ImageMetadataReader
import com.drew.metadata.exif.ExifIFD0Directory
import net.coobird.thumbnailator.Thumbnails
import net.cymoo.pebble.config.UploadConfig
import net.cymoo.pebble.exception.BadRequestException
import net.cymoo.pebble.model.FileInfo
import org.springframework.stereotype.Service
import org.springframework.web.multipart.MultipartFile
import java.awt.image.BufferedImage
import java.io.File
import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.Paths
import java.nio.file.StandardCopyOption
import java.util.*
import javax.imageio.ImageIO

@Service
class UploadService(private val uploadConfig: UploadConfig) {

    fun handleFileUpload(file: MultipartFile): FileInfo {
        if (file.isEmpty) {
            throw BadRequestException("File is empty")
        }

        if (file.originalFilename.isNullOrBlank()) {
            throw BadRequestException("Filename is required")
        }

        val filename =
            generateSecureFilename(file.originalFilename!!)
        val filepath = Paths.get(uploadConfig.uploadDir, filename)

        Files.copy(file.inputStream, filepath, StandardCopyOption.REPLACE_EXISTING)

        return when {
            isImage(file.contentType) -> processImageFile(filepath, file.contentType!!)
            else -> processRegularFile(filepath)
        }
    }

    private fun processRegularFile(filepath: Path): FileInfo {
        return FileInfo(
            url = "/${uploadConfig.uploadUrl}/${filepath.fileName}",
            size = filepath.toFile().length(),
            thumbUrl = null,
            width = null,
            height = null
        )
    }

    private fun isImage(contentType: String?): Boolean {
        return contentType?.removePrefix("image/") in uploadConfig.imageFormats
    }

    private fun processImageFile(filepath: Path, contentType: String): FileInfo {
        val file = filepath.toFile()
        val bufferedImage = ImageIO.read(file)

        // Handle image orientation based on EXIF
        val metadata = ImageMetadataReader.readMetadata(file)
        val exifDirectory = metadata.getFirstDirectoryOfType(ExifIFD0Directory::class.java)
        val orientation = if (exifDirectory?.containsTag(ExifIFD0Directory.TAG_ORIENTATION) == true) {
            exifDirectory.getInt(ExifIFD0Directory.TAG_ORIENTATION)
        } else {
            null
        }

        // Transpose image if needed
        val finalImage = when (orientation) {
            6 -> rotateImage(bufferedImage, 90)
            3 -> rotateImage(bufferedImage, 180)
            8 -> rotateImage(bufferedImage, 270)
            else -> bufferedImage
        }

        // Save transposed image if needed
        if (orientation != null && orientation != 1) {
            ImageIO.write(finalImage, contentType.removePrefix("image/"), file)
        }

        // Generate thumbnail for image
        val thumbUrl = runCatching { generateThumbnail(filepath, finalImage) }.getOrNull()

        return FileInfo(
            url = "/${uploadConfig.uploadUrl}/${filepath.fileName}",
            thumbUrl = thumbUrl?.let { "/${uploadConfig.uploadUrl}/${it.name}" },
            size = file.length(),
            width = finalImage.width,
            height = finalImage.height
        )
    }

    private fun rotateImage(bufferedImage: BufferedImage, degrees: Int): BufferedImage {
        return Thumbnails.of(bufferedImage)
            .scale(1.0)
            .rotate(degrees.toDouble())
            .asBufferedImage()
    }

    private fun generateThumbnail(originalPath: Path, image: BufferedImage): File {
        val (_, uploadDir, thumbnailSize) = uploadConfig

        return Paths.get(uploadDir, "thumb_${originalPath.fileName}").toFile().also {
            Thumbnails.of(image)
                .size(thumbnailSize, thumbnailSize)
                .keepAspectRatio(true)
                .toFile(it)
        }
    }
}

// Helper functions

val INVALID_CHARS_REGEX = Regex("[^\\w\\-.\\u4e00-\\u9fa5]+")

/**
 * Generates a secure filename with UUID suffix of specified length
 *
 * @param filename Original filename
 * @param uuidLength Length of UUID suffix (8-32 characters)
 * @return Secured filename with UUID suffix
 * @throws IllegalArgumentException if filename is blank or uuid length is invalid
 */
fun generateSecureFilename(filename: String, uuidLength: Int = 8): String {
    require(filename.isNotBlank()) { "Filename cannot be blank" }
    require(uuidLength in 8..32) {
        "UUID length must be between 8 and 32"
    }

    val sanitizedName = filename.trim().replace(INVALID_CHARS_REGEX, "_")

    val (baseName, extension) = splitFileName(sanitizedName)
    val uuid = UUID.randomUUID().toString()
        .replace("-", "")
        .take(uuidLength)

    return buildString {
        append(baseName)
        append('.')
        append(uuid)
        if (extension.isNotEmpty()) {
            append('.')
            append(extension)
        }
    }
}

/**
 * Splits a filename into base name and extension
 * Handles special cases like hidden files and multiple extensions
 */
fun splitFileName(fileName: String): Pair<String, String> {
    // Handle hidden file starting with `.`
    if (fileName.startsWith(".")) {
        val remaining = fileName.substring(1)
        val lastDotIndex = remaining.lastIndexOf('.')
        return if (lastDotIndex < 0) {
            ".$remaining" to ""
        } else {
            ".${remaining.take(lastDotIndex)}" to remaining.substring(lastDotIndex + 1)
        }
    }

    val lastDotIndex = fileName.lastIndexOf('.')
    return when {
        lastDotIndex <= 0 -> fileName to ""
        else -> fileName.take(lastDotIndex) to fileName.substring(lastDotIndex + 1)
    }
}
