
/// The Pipe trait provides a method to pipe a value through a transformation.
///
/// This trait allows for a more functional programming style by enabling
/// method chaining and easy value transformation.
///
/// # Examples
///
/// ```rust
/// use pebble::util::fp::Pipe;
/// let result = 5.pipe(|x| x * 2);  // result is 10
/// let string = "hello".pipe(|s| s.to_uppercase());  // string is "HELLO"
/// ```
pub trait Pipe {
    /// Transforms the current value by applying the given function.
    ///
    /// # Arguments
    ///
    /// * `f` - A closure that takes the current value and returns a transformed value
    ///
    /// # Returns
    ///
    /// The result of applying the transformation function to the current value
    fn pipe<F, R>(self, f: F) -> R
    where
        F: FnOnce(Self) -> R,
        Self: Sized;
}

impl<T> Pipe for T {
    fn pipe<F, R>(self, f: F) -> R
    where
        F: FnOnce(Self) -> R,
        Self: Sized,
    {
        // Apply the transformation function to the current value
        f(self)
    }
}
