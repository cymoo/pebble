package net.cymoo.pebble.util.maybe;

import jakarta.validation.valueextraction.ExtractedValue;
import jakarta.validation.valueextraction.UnwrapByDefault;
import jakarta.validation.valueextraction.ValueExtractor;

/**
 * Extractor for MaybeMissing
 */
@UnwrapByDefault
public class MaybeMissingValueExtractor implements ValueExtractor<MaybeMissing<@ExtractedValue ?>> {
    @Override
    public void extractValues(MaybeMissing<?> originalValue, ValueReceiver receiver) {
        if (originalValue.isPresent()) {
            receiver.value(null, originalValue.get());
        }
    }
}
