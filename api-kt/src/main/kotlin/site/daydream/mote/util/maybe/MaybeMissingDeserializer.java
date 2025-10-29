package site.daydream.mote.util.maybe;

import com.fasterxml.jackson.databind.*;
import com.fasterxml.jackson.databind.deser.Deserializers;
import com.fasterxml.jackson.databind.deser.ValueInstantiator;
import com.fasterxml.jackson.databind.deser.std.ReferenceTypeDeserializer;
import com.fasterxml.jackson.databind.jsontype.TypeDeserializer;
import com.fasterxml.jackson.databind.type.ReferenceType;

class MaybeMissingDeserializer extends ReferenceTypeDeserializer<MaybeMissing<?>> {
    public MaybeMissingDeserializer(JavaType fullType, ValueInstantiator inst, TypeDeserializer typeDeser, JsonDeserializer<?> deser) {
        super(fullType, inst, typeDeser, deser);
    }

    @Override
    public MaybeMissingDeserializer withResolved(TypeDeserializer typeDeser, JsonDeserializer<?> valueDeser) {
        return new MaybeMissingDeserializer(_fullType, _valueInstantiator, typeDeser, valueDeser);
    }

    @Override
    public MaybeMissing<?> getNullValue(DeserializationContext ctxt) {
        return MaybeMissing.of(null);
    }

    @Override
    public MaybeMissing<?> getAbsentValue(DeserializationContext ctxt) {
        return MaybeMissing.missing();
    }

    @Override
    public MaybeMissing<?> referenceValue(Object contents) {
        return MaybeMissing.of(contents);
    }

    @Override
    public MaybeMissing<?> updateReference(MaybeMissing<?> reference, Object contents) {
        return referenceValue(contents);
    }

    @Override
    public Object getReferenced(MaybeMissing<?> reference) {
        return reference.get();
    }
}

class MaybeMissingDeserializers extends Deserializers.Base {
    @Override
    public JsonDeserializer<?> findReferenceDeserializer(ReferenceType refType, DeserializationConfig config,
                                                         BeanDescription beanDesc, TypeDeserializer contentTypeDeserializer,
                                                         JsonDeserializer<?> contentDeserializer) {
        if (refType.hasRawClass(MaybeMissing.class)) {
            return new MaybeMissingDeserializer(refType, null, contentTypeDeserializer, contentDeserializer);
        }
        return null;
    }
}

