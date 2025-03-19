package net.cymoo.pebble.annotation

@Target(AnnotationTarget.FUNCTION, AnnotationTarget.CLASS)
@Retention(AnnotationRetention.RUNTIME)
annotation class AuthRequired(val required: Boolean = true)
