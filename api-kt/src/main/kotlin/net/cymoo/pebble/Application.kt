package net.cymoo.pebble

import net.cymoo.pebble.util.Env
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.context.properties.ConfigurationPropertiesScan
import org.springframework.boot.runApplication
import org.springframework.context.annotation.EnableAspectJAutoProxy
import org.springframework.scheduling.annotation.EnableAsync
import org.springframework.scheduling.annotation.EnableScheduling
import org.springframework.transaction.annotation.EnableTransactionManagement
import kotlin.system.exitProcess

@SpringBootApplication
@EnableAspectJAutoProxy
@EnableScheduling
@EnableAsync
@EnableTransactionManagement
@ConfigurationPropertiesScan
class SpringApplication

fun main(args: Array<String>) {
    Env.load()

    if (Env.get("PEBBLE_PASSWORD").isNullOrBlank()) {
        System.err.println("Error: PEBBLE_PASSWORD environment variable is missing.")
        exitProcess(1)
    }
    runApplication<SpringApplication>(*args)
}

val <T : Any> T.logger: Logger
    get() = LoggerFactory.getLogger(this::class.java)

