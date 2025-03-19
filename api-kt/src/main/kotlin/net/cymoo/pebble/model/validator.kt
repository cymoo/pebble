package net.cymoo.pebble.model

import jakarta.validation.Constraint
import jakarta.validation.ConstraintValidator
import jakarta.validation.ConstraintValidatorContext
import jakarta.validation.Payload
import java.time.LocalDate
import java.time.format.DateTimeFormatter.ISO_LOCAL_DATE
import kotlin.reflect.KClass

@Target(AnnotationTarget.FIELD)
@Retention(AnnotationRetention.RUNTIME)
@Constraint(validatedBy = [DateValidator::class])
annotation class ValidDate(
    val message: String = "must be in 'yyyy-MM-dd' format",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

class DateValidator : ConstraintValidator<ValidDate, String> {
    override fun isValid(value: String?, context: ConstraintValidatorContext?) =
        value == null || runCatching { LocalDate.parse(value, ISO_LOCAL_DATE) }.isSuccess
}
