use chrono::NaiveDate;
use validator::ValidationError;

/// Validate date string in "yyyy-MM-dd" format
pub fn validate_date_format(date: &str) -> Result<(), ValidationError> {
    if NaiveDate::parse_from_str(date, "%Y-%m-%d").is_ok() {
        Ok(())
    } else {
        let mut error = ValidationError::new("invalid_date");
        error.message = Some("must be in 'yyyy-MM-dd' format".into());
        Err(error)
    }
}
