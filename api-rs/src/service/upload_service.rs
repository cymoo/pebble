use crate::config::UploadConfig;
use crate::errors::{ApiError, ApiResult};
use crate::model::post::FileInfo;
use anyhow::{anyhow, Context, Result};
use axum::extract::multipart::{Field, MultipartError};
use exif::{In, Reader, Tag};
use futures_util::TryStreamExt;
use image::DynamicImage;
use image::ImageReader;
use regex::Regex;
use std::borrow::Cow;
use std::io;
use std::io::Cursor;
use std::path::{Path, PathBuf};
use tokio::fs;
use tokio::fs::File;
use tokio::io::BufWriter;
use tokio_util::io::StreamReader;
use tracing::error;
use uuid::Uuid;

pub struct FileUploadService {
    config: UploadConfig,
}

impl FileUploadService {
    pub fn new(config: UploadConfig) -> Self {
        Self { config }
    }

    pub async fn stream_to_file(&self, field: Field<'_>) -> ApiResult<FileInfo> {
        let file_name = field
            .file_name()
            .ok_or(ApiError::BadRequest("Invalid filename".into()))?;
        let content_type = field
            .content_type()
            .ok_or(ApiError::BadRequest("Invalid file type".into()))?
            .to_owned();

        let file_name = generate_secure_filename(file_name, 8);
        let upload_dir = self.config.base_path.clone();
        let file_path = Path::new(&upload_dir).join(file_name);

        // Convert the stream into an `AsyncRead`.
        let body_with_io_error = field.map_err(|err| io::Error::new(io::ErrorKind::Other, err));
        let body_reader = StreamReader::new(body_with_io_error);
        futures::pin_mut!(body_reader);

        let file = File::create(&file_path)
            .await
            .context("Cannot create file")?;
        let mut buf_writer = BufWriter::new(file);

        // Copy the body into the file.
        tokio::io::copy(&mut body_reader, &mut buf_writer)
            .await
            .map_err(|err| {
                std::fs::remove_file(&file_path)
                    .map_err(|e| error!("Cannot remove file: {}", e))
                    .ok();

                if let Ok(err) = err.downcast::<MultipartError>() {
                    ApiError::MultiPartError(err)
                } else {
                    ApiError::Anyhow(anyhow!("cannot save file"))
                }
            })?;

        if self.is_image(&content_type) {
            self.process_image_file(&file_path, &content_type)
                .await
                .map_err(ApiError::from)
        } else {
            self.process_regular_file(&file_path)
                .await
                .map_err(ApiError::from)
        }
    }

    async fn process_regular_file(&self, filepath: &Path) -> Result<FileInfo> {
        let metadata = fs::metadata(filepath).await?;
        Ok(FileInfo {
            url: format!("{}/{}", self.config.base_url, Self::get_filename(filepath)),
            size: Some(metadata.len()),
            thumb_url: None,
            width: None,
            height: None,
        })
    }

    async fn process_image_file(&self, filepath: &Path, content_type: &str) -> Result<FileInfo> {
        // Read Image
        let bytes = tokio::fs::read(filepath).await?;

        let img = if Self::needs_exif_rotation(content_type) {
            self.handle_exif_rotation(&bytes, filepath)?
        } else {
            ImageReader::new(Cursor::new(&bytes))
                .with_guessed_format()?
                .decode()?
        };

        // Generate thumbnail
        let thumb_path = self
            .generate_thumbnail(filepath, &img)
            .context("Cannot create thumbnail")?;

        let thumb_url = format!(
            "{}/{}",
            self.config.base_url,
            Self::get_filename(&thumb_path)
        );

        let url = format!("{}/{}", self.config.base_url, Self::get_filename(filepath));

        // Read filesize
        let metadata = fs::metadata(filepath).await?;

        Ok(FileInfo {
            url,
            thumb_url: Some(thumb_url),
            size: Some(metadata.len()),
            width: Some(img.width()),
            height: Some(img.height()),
        })
    }

    fn handle_exif_rotation(&self, bytes: &Vec<u8>, path: &Path) -> Result<DynamicImage> {
        let mut cursor = Cursor::new(&bytes);

        let exif = Reader::new().read_from_container(&mut cursor).ok();

        // Handle exif orientation
        let orientation = exif.and_then(|exif| {
            exif.get_field(Tag::Orientation, In::PRIMARY)
                .and_then(|field| field.value.get_uint(0))
        });

        cursor.set_position(0);

        let mut img = ImageReader::new(cursor).with_guessed_format()?.decode()?;

        img = match orientation {
            Some(6) => img.rotate90(),
            Some(3) => img.rotate180(),
            Some(8) => img.rotate270(),
            _ => img,
        };

        if orientation.is_some() {
            img.save(path)?;
        }
        Ok(img)
    }

    fn needs_exif_rotation(content_type: &str) -> bool {
        let format = content_type
            .strip_prefix("image/")
            .unwrap_or("")
            .to_string()
            .to_lowercase();
        match format.as_str() {
            "jpeg" | "jpg" | "jif" => true,
            "png" | "gif" | "webp" | "avif" | "svg" => false,
            _ => false,
        }
    }

    fn generate_thumbnail(&self, original_path: &Path, img: &DynamicImage) -> Result<PathBuf> {
        let thumb_filename = format!("thumb_{}", Self::get_filename(original_path));
        let thumb_path = PathBuf::from(&self.config.base_path).join(&thumb_filename);

        let thumbnail = img.thumbnail(self.config.thumb_width, self.config.thumb_width);

        thumbnail.save(&thumb_path)?;
        Ok(thumb_path)
    }

    fn is_image(&self, content_type: &str) -> bool {
        let format = content_type
            .strip_prefix("image/")
            .unwrap_or("")
            .to_string()
            .to_lowercase();
        self.config.image_formats.contains(&format)
    }

    fn get_filename(filepath: &Path) -> Cow<'_, str> {
        filepath.file_name().unwrap().to_string_lossy()
    }
}

// Helper functions

/// Generates a secure filename by sanitizing the input filename and appending a UUID.
///
/// # Arguments
/// * `filename` - The original filename to be sanitized.
/// * `uuid_length` - The length of the UUID to be appended (must be between 8 and 32).
///
/// # Returns
/// A `String` representing the secure filename in the format: `basename.uuid` or `basename.uuid.extension`.
///
/// # Panics
/// - Panics if the `filename` is empty.
/// - Panics if the `uuid_length` is not between 8 and 32.
pub fn generate_secure_filename(filename: &str, uuid_length: usize) -> String {
    if filename.is_empty() {
        panic!("filename is empty");
    }

    if !(8..=32).contains(&uuid_length) {
        panic!("uuid length must be between 8 and 32");
    }

    let invalid_chars_regex = Regex::new(r"[^\w\-.\u4e00-\u9fa5]+").unwrap();
    let sanitized_name = invalid_chars_regex.replace_all(filename.trim(), "_");

    let (base, ext) = split_filename(&sanitized_name);

    let uuid = Uuid::new_v4()
        .to_string()
        .chars()
        .take(uuid_length)
        .collect::<String>();

    if ext.is_empty() {
        format!("{}.{}", base, uuid)
    } else {
        format!("{}.{}.{}", base, uuid, ext)
    }
}

/// Splits a filename into its base name and extension.
///
/// # Arguments
/// * `filename` - The filename to be split.
///
/// # Returns
/// A tuple `(base_name, extension)` where:
/// - `base_name` is the part of the filename before the last dot.
/// - `extension` is the part of the filename after the last dot, converted to lowercase.
///   If the filename has no extension, `extension` will be an empty string.
///
/// # Panics
/// - Panics if the `filename` is empty.
pub fn split_filename(filename: &str) -> (String, String) {
    if filename.is_empty() {
        panic!("filename is empty");
    }

    if let Some(remainder) = filename.strip_prefix('.') {
        match remainder.rfind('.') {
            Some(i) => (
                format!(".{}", &remainder[..i]),
                remainder[i + 1..].to_string().to_lowercase(),
            ),
            None => (filename.to_string(), String::new()),
        }
    } else {
        match filename.rfind('.') {
            Some(i) => (
                filename[..i].to_string(),
                filename[i + 1..].to_string().to_lowercase(),
            ),
            None => (filename.to_string(), String::new()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_split_filename() {
        let filenames = vec![
            ("foo/bar.jpg", "foo/bar", "jpg"),
            ("foo.jpg", "foo", "jpg"),
            ("foo.JPG", "foo", "jpg"),
            (".foo.jpg", ".foo", "jpg"),
            (".Foo.JPG", ".Foo", "jpg"),
            (".foo", ".foo", ""),
            (".foo.tar.gz", ".foo.tar", "gz"),
            ("foo", "foo", ""),
        ];

        filenames.into_iter().for_each(|(input, name, ext)| {
            assert_eq!(split_filename(input), (name.to_string(), ext.to_string()));
        });
    }
}
