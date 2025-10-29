use crate::errors::ApiError;
use axum::extract::rejection::PathRejection;
use axum::extract::{FromRequest, FromRequestParts, Request};
use axum::http::request::Parts;
use axum::response::{IntoResponse, Response};
use serde::de::DeserializeOwned;
use serde::Serialize;
use validator::Validate;

/// an extractor that internally uses `axum::extract::Json` but has a custom rejection
#[derive(FromRequest)]
#[from_request(via(axum::extract::Json), rejection(ApiError))]
pub struct Json<T>(pub T);

impl<T: Serialize> IntoResponse for Json<T> {
    fn into_response(self) -> Response {
        let Self(value) = self;
        axum::Json(value).into_response()
    }
}

/// an extractor that internally uses `axum::extract::Form` but has a custom rejection
#[derive(FromRequest)]
#[from_request(via(axum::extract::Form), rejection(ApiError))]
pub struct Form<T>(pub T);

/// an extractor that internally uses `axum::extract::Query` but has a custom rejection
#[derive(FromRequestParts)]
#[from_request(via(axum::extract::Query), rejection(ApiError))]
pub struct Query<T>(pub T);

/// an extractor that internally uses `crate::extractor::wrapper::Query` and adds validation
#[derive(Debug, Clone, Copy, Default)]
pub struct ValidatedQuery<T>(pub T);

impl<T, S> FromRequestParts<S> for ValidatedQuery<T>
where
    T: DeserializeOwned + Validate,
    S: Send + Sync,
    Query<T>: FromRequestParts<S, Rejection = ApiError>,
{
    type Rejection = ApiError;

    async fn from_request_parts(parts: &mut Parts, state: &S) -> Result<Self, Self::Rejection> {
        let Query(value) = Query::<T>::from_request_parts(parts, state).await?;
        value.validate()?;
        Ok(ValidatedQuery(value))
    }
}

/// an extractor that internally uses `crate::extractor::wrapper::Form` and adds validation
#[derive(Debug, Clone, Copy, Default)]
pub struct ValidatedForm<T>(pub T);

impl<T, S> FromRequest<S> for ValidatedForm<T>
where
    T: DeserializeOwned + Validate,
    S: Send + Sync,
    Form<T>: FromRequest<S, Rejection = ApiError>,
{
    type Rejection = ApiError;

    async fn from_request(req: Request, state: &S) -> Result<Self, Self::Rejection> {
        let Form(value) = Form::<T>::from_request(req, state).await?;
        value.validate()?;
        Ok(ValidatedForm(value))
    }
}

/// an extractor that internally uses `crate::extractor::wrapper::Json` and adds validation
#[derive(Debug, Clone, Copy, Default)]
pub struct ValidatedJson<T>(pub T);

impl<T, S> FromRequest<S> for ValidatedJson<T>
where
    T: DeserializeOwned + Validate,
    S: Send + Sync,
    Json<T>: FromRequest<S, Rejection = ApiError>,
{
    type Rejection = ApiError;

    async fn from_request(req: Request, state: &S) -> Result<Self, Self::Rejection> {
        let Json(value) = Json::<T>::from_request(req, state).await?;
        value.validate()?;
        Ok(ValidatedJson(value))
    }
}

/// an extractor that internally uses `axum::extract::Path` but has a custom rejection
pub struct Path<T>(pub T);

impl<S, T> FromRequestParts<S> for Path<T>
where
    // these trait bounds are copied from `impl FromRequest for axum::extract::path::Path`
    T: DeserializeOwned + Send,
    S: Send + Sync,
{
    type Rejection = ApiError;

    async fn from_request_parts(parts: &mut Parts, state: &S) -> Result<Self, Self::Rejection> {
        use axum::extract::path::ErrorKind::*;
        use ApiError::PathError;

        match axum::extract::Path::<T>::from_request_parts(parts, state).await {
            Ok(value) => Ok(Self(value.0)),
            Err(rejection) => {
                Err(match rejection {
                    PathRejection::FailedToDeserializePathParams(inner) => {
                        let kind = inner.into_kind();

                        match &kind {
                            WrongNumberOfParameters { .. }
                            | ParseErrorAtKey { .. }
                            | ParseErrorAtIndex { .. }
                            | ParseError { .. }
                            | InvalidUtf8InPathParam { .. } => PathError(400, kind.to_string()),

                            UnsupportedType { .. } => {
                                // this error is caused by the programmer using an unsupported type
                                // (such as nested maps) so respond with `500` instead
                                PathError(500, kind.to_string())
                            }

                            Message(msg) => PathError(400, msg.to_string()),
                            _ => PathError(400, format!("Unhandled deserialization error: {kind}")),
                        }
                    }
                    PathRejection::MissingPathParams(error) => PathError(500, error.to_string()),
                    _ => PathError(500, format!("Unhandled path rejection: {rejection}")),
                })
            }
        }
    }
}
