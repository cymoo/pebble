package net.cymoo.pebble.util.maybe;

import com.fasterxml.jackson.databind.JavaType;
import com.fasterxml.jackson.databind.type.ReferenceType;
import com.fasterxml.jackson.databind.type.TypeBindings;
import com.fasterxml.jackson.databind.type.TypeFactory;
import com.fasterxml.jackson.databind.type.TypeModifier;

import java.lang.reflect.Type;

class MaybeMissingTypeModifier extends TypeModifier {
    @Override
    public JavaType modifyType(JavaType type, Type jdkType, TypeBindings context, TypeFactory typeFactory) {
        if (type.isReferenceType() || type.isContainerType()) {
            return type;
        }

        if (type.getRawClass() == MaybeMissing.class) {
            return ReferenceType.upgradeFrom(type, type.containedTypeOrUnknown(0));
        }
        return type;
    }
}


