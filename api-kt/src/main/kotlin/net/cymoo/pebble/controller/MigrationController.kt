package net.cymoo.pebble.controller

import net.cymoo.pebble.service.MigrationInfo
import net.cymoo.pebble.service.MigrationService
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/migrations")
class MigrationController(private val migrationService: MigrationService) {

    @GetMapping
    fun getMigrationInfo(): List<MigrationInfo> {
        return migrationService.getMigrationInfo()
    }

    @PostMapping("/repair")
    fun repair() {
        migrationService.repair()
    }

    @PostMapping("/migrate")
    fun migrate() {
        migrationService.migrate()
    }
}
