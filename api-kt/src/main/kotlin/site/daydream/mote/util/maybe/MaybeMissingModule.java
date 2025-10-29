package site.daydream.mote.util.maybe;

import com.fasterxml.jackson.core.Version;
import com.fasterxml.jackson.databind.Module;


public class MaybeMissingModule extends Module {
    @Override
    public String getModuleName() {
        return "MaybeMissingModule";
    }

    @Override
    public Version version() {
        return Version.unknownVersion();
    }

    @Override
    public void setupModule(Module.SetupContext context) {
        context.addDeserializers(new MaybeMissingDeserializers());
        context.addTypeModifier(new MaybeMissingTypeModifier());
    }
}
