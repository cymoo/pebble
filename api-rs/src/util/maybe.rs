use serde::{Deserialize, Serialize};
use validator::{Validate, ValidationErrors};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(untagged)]
pub enum MaybeAbsent<T> {
    Present(T),
    #[serde(skip_serializing)]
    Absent,
}

impl<T> Default for MaybeAbsent<T> {
    fn default() -> Self {
        Self::Absent
    }
}

impl<T> MaybeAbsent<T> {
    pub fn is_absent(&self) -> bool {
        matches!(self, Self::Absent)
    }

    pub fn is_present(&self) -> bool {
        !self.is_absent()
    }

    pub fn as_ref(&self) -> MaybeAbsent<&T> {
        match self {
            MaybeAbsent::Present(v) => MaybeAbsent::Present(v),
            MaybeAbsent::Absent => MaybeAbsent::Absent,
        }
    }

    pub fn get(&self) -> &T {
        match self {
            MaybeAbsent::Present(v) => v,
            MaybeAbsent::Absent => panic!("Value is absent!"),
        }
    }

    pub fn get_or<'a>(&'a self, default: &'a T) -> &'a T {
        match self {
            MaybeAbsent::Present(v) => v,
            MaybeAbsent::Absent => default,
        }
    }

    pub fn if_present<'a, F>(&'a self, f: F)
    where
        F: FnOnce(&'a T),
    {
        if let MaybeAbsent::Present(v) = self {
            f(v);
        }
    }

    pub fn map<U, F>(self, f: F) -> MaybeAbsent<U>
    where
        F: FnOnce(T) -> U,
    {
        match self {
            MaybeAbsent::Present(value) => MaybeAbsent::Present(f(value)),
            MaybeAbsent::Absent => MaybeAbsent::Absent,
        }
    }
}

// TODO: it do not work
impl<T: Validate> Validate for MaybeAbsent<T> {
    fn validate(&self) -> Result<(), ValidationErrors> {
        match self {
            MaybeAbsent::Present(value) => value.validate(),
            MaybeAbsent::Absent => Ok(()),
        }
    }
}
