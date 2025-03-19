package net.cymoo.pebble.util.maybe;

import java.util.Objects;
import java.util.function.Consumer;

// https://github.com/OpenAPITools/jackson-databind-nullable
// https://stackoverflow.com/questions/55166379/deserialize-generic-type-using-referencetypedeserializer-with-jackson-spring
public class MaybeMissing<T> {
    private static final MaybeMissing<?> MISSING = new MaybeMissing<>(null, false);
    private final T value;
    private final boolean isPresent;

    private MaybeMissing(T value, boolean isPresent) {
        this.value = value;
        this.isPresent = isPresent;
    }

    public static <T> MaybeMissing<T> of(T value) {
        return new MaybeMissing<>(value, true);
    }

    @SuppressWarnings("unchecked")
    public static <T> MaybeMissing<T> missing() {
        return (MaybeMissing<T>) MISSING;
    }

    public boolean isPresent() {
        return isPresent;
    }

    public T get() {
        if (!isPresent) {
            throw new IllegalStateException("Value is missing");
        }
        return value;
    }

    public T getOrElse(T defaultValue) {
        return isPresent() ? this.get() : defaultValue;
    }

    public void ifPresent(Consumer<T> consumer) {
        if (isPresent()) {
            consumer.accept(this.get());
        }
    }

    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof MaybeMissing<?> other)) return false;
        return Objects.equals(value, other.value) && isPresent == other.isPresent;
    }

    @Override
    public int hashCode() {
        return Objects.hash(value, isPresent);
    }

    @Override
    public String toString() {
        if (isPresent) {
            return String.format("Present[%s]", value);
        } else {
            return "Missing";
        }
    }
}

