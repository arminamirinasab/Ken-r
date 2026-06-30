package ir.kenar.domain.widget

import java.net.URLDecoder
import java.net.URLEncoder
import kotlin.math.abs

/**
 * Snapshot for a shared countdown widget.
 *
 * Widgets stay passive: the sync/data layer calculates [remainingDays] for the
 * user's active locale/calendar day and writes the snapshot into Glance state.
 */
data class Countdown(
    val title: String,
    val targetEpochDay: Long,
    val remainingDays: Int,
    val asOfEpochDay: Long,
    val targetLabel: String = "",
) {
    init {
        require(title.isValidText(MAX_TITLE_CHARS)) { "invalid title" }
        require(targetEpochDay in MIN_EPOCH_DAY..MAX_EPOCH_DAY) { "targetEpochDay out of range" }
        require(asOfEpochDay in MIN_EPOCH_DAY..MAX_EPOCH_DAY) { "asOfEpochDay out of range" }
        require(abs(remainingDays) <= MAX_REMAINING_DAYS) { "remainingDays out of range" }
        require(targetLabel.isValidText(MAX_TARGET_LABEL_CHARS, allowBlank = true)) { "invalid targetLabel" }
    }

    val state: CountdownState
        get() = when {
            remainingDays > 0 -> CountdownState.UPCOMING
            remainingDays == 0 -> CountdownState.TODAY
            else -> CountdownState.PASSED
        }

    fun encodeWire(): String {
        val wire = listOf(
            WIRE_VERSION,
            title.escape(),
            targetEpochDay.toString(),
            remainingDays.toString(),
            asOfEpochDay.toString(),
            targetLabel.escape(),
        ).joinToString(SECTION_SEPARATOR)

        require(wire.length <= MAX_WIRE_CHARS) { "countdown payload too large" }
        return wire
    }

    companion object {
        const val MAX_TITLE_CHARS = 48
        const val MAX_TARGET_LABEL_CHARS = 32
        const val MAX_REMAINING_DAYS = 36500
        const val MAX_WIRE_CHARS = 512

        // Roughly +/- 100 years around Unix epoch. Wide enough for app events,
        // tight enough to reject accidental milliseconds.
        private const val MIN_EPOCH_DAY = -36525L
        private const val MAX_EPOCH_DAY = 36525L
        private const val WIRE_VERSION = "v1"
        private const val SECTION_SEPARATOR = "|"

        fun fromWire(value: String?): Countdown? {
            if (value.isNullOrBlank() || value.length > MAX_WIRE_CHARS) return null

            return runCatching {
                val parts = value.split(SECTION_SEPARATOR)
                if (parts.size != 6 || parts[0] != WIRE_VERSION) return null

                val title = parts[1].unescape() ?: return null
                val targetEpochDay = parts[2].toLongOrNull() ?: return null
                val remainingDays = parts[3].toIntOrNull() ?: return null
                val asOfEpochDay = parts[4].toLongOrNull() ?: return null
                val targetLabel = parts[5].unescape() ?: return null

                Countdown(
                    title = title,
                    targetEpochDay = targetEpochDay,
                    remainingDays = remainingDays,
                    asOfEpochDay = asOfEpochDay,
                    targetLabel = targetLabel,
                )
            }.getOrNull()
        }

        private fun String.escape(): String = URLEncoder.encode(this, "UTF-8")

        private fun String.unescape(): String? = runCatching {
            URLDecoder.decode(this, "UTF-8")
        }.getOrNull()

        private fun String.isValidText(maxChars: Int, allowBlank: Boolean = false): Boolean {
            if (!allowBlank && isBlank()) return false
            if (length > maxChars) return false
            return none { it.code < 32 || it.code == 127 }
        }
    }
}

enum class CountdownState {
    UPCOMING,
    TODAY,
    PASSED,
}
