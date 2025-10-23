package net.cymoo.pebble.service

import com.drew.imaging.ImageMetadataReader
import com.drew.metadata.exif.ExifIFD0Directory
import net.coobird.thumbnailator.Thumbnails
import net.cymoo.pebble.config.FileUploadConfig
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
import javax.imageio.ImageIO

@Service
class FileUploadService(private val uploadConfig: FileUploadConfig) {

    fun handleFileUpload(file: MultipartFile): FileInfo {
        if (file.isEmpty) {
            throw BadRequestException("File is empty")
        }

        if (file.originalFilename.isNullOrBlank()) {
            throw BadRequestException("Filename is required")
        }

        val filename =
            net.cymoo.pebble.util.generateSecureFilename(file.originalFilename!!)
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
        // TODO: upload PNG will fail:
        // Tag 'Orientation' has not been set -- check using containsTag() first
        val orientation = exifDirectory?.getInt(ExifIFD0Directory.TAG_ORIENTATION)

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


