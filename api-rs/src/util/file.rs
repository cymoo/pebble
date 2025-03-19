use regex::Regex;
use std::path::Path;
use uuid::Uuid;

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

    if uuid_length < 8 || uuid_length > 32 {
        panic!("uuid length must be between 8 and 32");
    }

    let invalid_chars_regex = Regex::new(r"[^\w\-.\u4e00-\u9fa5]+").unwrap();
    let sanitized_name = invalid_chars_regex.replace_all(&filename.trim(), "_");

    let (base, ext) = split_filename(&*sanitized_name);

    let uuid = Uuid::new_v4().to_string()
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

    if filename.starts_with('.') {
        let remaining = &filename[1..];
        match remaining.rfind('.') {
            Some(i) => (
                format!(".{}", &remaining[..i]),
                remaining[i + 1..].to_string().to_lowercase(),
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

/// Checks if a given path is valid and consists of a single, normal component.
/// Prevent directory traversal attacks we ensure the path consists of exactly one normal component.
///
/// A path is considered valid if:
/// - It contains only one component.
/// - Not a prefix, root directory, parent directory, or current directory.
///
/// # Arguments
/// * `path` - The path to be validated.
///
/// # Returns
/// - `true` if the path is valid and consists of a single normal component.
/// - `false` otherwise.
///
/// # Examples
/// ```
/// use pebble::util::file::path_is_valid;
/// assert_eq!(path_is_valid("file.txt"), true);
/// assert_eq!(path_is_valid("folder/file.txt"), false); // Multiple components
/// assert_eq!(path_is_valid(".."), false);             // Parent directory
/// assert_eq!(path_is_valid("."), false);              // Current directory
/// assert_eq!(path_is_valid("/root"), false);          // Root directory
/// ```
pub fn path_is_valid(path: &str) -> bool {
    let path = Path::new(path);
    let mut components = path.components().peekable();

    if let Some(first) = components.peek() {
        if !matches!(first, std::path::Component::Normal(_)) {
            return false;
        }
    }

    components.count() == 1
}

pub fn parent_path(path: &str) -> String {
    let path = Path::new(path);
    match path.parent() {
        Some(parent) => parent.to_str().unwrap_or("").to_string(),
        None => String::new(),
    }
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parent_path() {
        assert_eq!(parent_path("a"), "");
        assert_eq!(parent_path("a/b"), "a");
        assert_eq!(parent_path("a/b/c"), "a/b");
    }

    #[test]
    fn test_path_is_valid() {
        assert!(path_is_valid("foo"));
        assert!(path_is_valid("foo.jpg"));
        assert!(!path_is_valid("./foo.jpg"));
        assert!(!path_is_valid("/foo.jpg"));
        assert!(!path_is_valid("/foo/bar.jpg"));
    }

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
