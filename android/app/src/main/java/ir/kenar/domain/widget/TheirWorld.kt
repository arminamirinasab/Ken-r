package ir.kenar.domain.widget

import androidx.annotation.StringRes
import ir.kenar.R
import java.net.URLDecoder
import java.net.URLEncoder

/**
 * Snapshot of the partner's surrounding world: local time, weather, and day/night.
 *
 * The sync layer owns location permission, weather lookup, timezone conversion,
 * and privacy policy. The widget receives only a compact snapshot and renders it
 * passively.
 */
data class TheirWorld(
    val locationLabel: String,
    val localTimeLabel: String,
    val utcOffsetMinutes: Int,
    val weather: WeatherCondition,
    val temperatureCelsius: Int?,
    val dayNight: DayNightState,
    val observedAtEpochSeconds: Long,
) {
    init {
        require(locationLabel.isValidText(MAX_LOCATION_CHARS)) { "invalid locationLabel" }
        require(localTimeLabel.isValidText(MAX_TIME_LABEL_CHARS)) { "invalid localTimeLabel" }
        require(utcOffsetMinutes in MIN_UTC_OFFSET_MINUTES..MAX_UTC_OFFSET_MINUTES) {
            "utcOffsetMinutes out of range"
        }
        require(temperatureCelsius == null || temperatureCelsius in MIN_TEMP_CELSIUS..MAX_TEMP_CELSIUS) {
            "temperatureCelsius out of range"
        }
        require(observedAtEpochSeconds >= 0) { "observedAtEpochSeconds out of range" }
    }

    fun encodeWire(): String {
        val temp = temperatureCelsius?.toString().orEmpty()
        val wire = listOf(
            WIRE_VERSION,
            locationLabel.escape(),
            localTimeLabel.escape(),
            utcOffsetMinutes.toString(),
            weather.wireValue,
            temp,
            dayNight.wireValue,
            observedAtEpochSeconds.toString(),
        ).joinToString(SECTION_SEPARATOR)

        require(wire.length <= MAX_WIRE_CHARS) { "their world payload too large" }
        return wire
    }

    companion object {
        const val MAX_LOCATION_CHARS = 64
        const val MAX_TIME_LABEL_CHARS = 16
        const val MAX_WIRE_CHARS = 1024

        private const val MIN_UTC_OFFSET_MINUTES = -14 * 60
        private const val MAX_UTC_OFFSET_MINUTES = 14 * 60
        private const val MIN_TEMP_CELSIUS = -80
        private const val MAX_TEMP_CELSIUS = 80
        private const val WIRE_VERSION = "v1"
        private const val SECTION_SEPARATOR = "|"

        fun fromWire(value: String?): TheirWorld? {
            if (value.isNullOrBlank() || value.length > MAX_WIRE_CHARS) return null

            return runCatching {
                val parts = value.split(SECTION_SEPARATOR)
                if (parts.size != 8 || parts[0] != WIRE_VERSION) return null

                val location = parts[1].unescape() ?: return null
                val localTime = parts[2].unescape() ?: return null
                val offset = parts[3].toIntOrNull() ?: return null
                val weather = WeatherCondition.fromWire(parts[4]) ?: return null
                val temp = if (parts[5].isBlank()) {
                    null
                } else {
                    parts[5].toIntOrNull() ?: return null
                }
                val dayNight = DayNightState.fromWire(parts[6]) ?: return null
                val observedAt = parts[7].toLongOrNull() ?: return null

                TheirWorld(
                    locationLabel = location,
                    localTimeLabel = localTime,
                    utcOffsetMinutes = offset,
                    weather = weather,
                    temperatureCelsius = temp,
                    dayNight = dayNight,
                    observedAtEpochSeconds = observedAt,
                )
            }.getOrNull()
        }

        private fun String.escape(): String = URLEncoder.encode(this, "UTF-8")

        private fun String.unescape(): String? = runCatching {
            URLDecoder.decode(this, "UTF-8")
        }.getOrNull()

        private fun String.isValidText(maxChars: Int): Boolean =
            isNotBlank() && length <= maxChars && none { it.code < 32 || it.code == 127 }
    }
}

enum class WeatherCondition(
    val wireValue: String,
    @StringRes val labelRes: Int,
    val glyph: String,
) {
    CLEAR("clear", R.string.world_weather_clear, "☀"),
    CLOUDY("cloudy", R.string.world_weather_cloudy, "☁"),
    RAIN("rain", R.string.world_weather_rain, "☂"),
    SNOW("snow", R.string.world_weather_snow, "❄"),
    STORM("storm", R.string.world_weather_storm, "⚡"),
    FOG("fog", R.string.world_weather_fog, "≋"),
    WINDY("windy", R.string.world_weather_windy, "〰"),
    UNKNOWN("unknown", R.string.world_weather_unknown, "•");

    companion object {
        fun fromWire(value: String?): WeatherCondition? = entries.firstOrNull { it.wireValue == value }
    }
}

enum class DayNightState(
    val wireValue: String,
    @StringRes val labelRes: Int,
) {
    DAY("day", R.string.world_day),
    NIGHT("night", R.string.world_night);

    companion object {
        fun fromWire(value: String?): DayNightState? = entries.firstOrNull { it.wireValue == value }
    }
}
