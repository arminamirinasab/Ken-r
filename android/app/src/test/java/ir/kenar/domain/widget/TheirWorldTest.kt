package ir.kenar.domain.widget

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TheirWorldTest {

    @Test
    fun encodeWire_roundTripsTheirWorldSnapshot() {
        val world = TheirWorld(
            locationLabel = "Tehran",
            localTimeLabel = "22:14",
            utcOffsetMinutes = 210,
            weather = WeatherCondition.CLEAR,
            temperatureCelsius = 27,
            dayNight = DayNightState.NIGHT,
            observedAtEpochSeconds = 1_783_000_000,
        )

        assertEquals(world, TheirWorld.fromWire(world.encodeWire()))
    }

    @Test
    fun fromWire_returnsNullForMalformedPayload() {
        assertNull(TheirWorld.fromWire(null))
        assertNull(TheirWorld.fromWire(""))
        assertNull(TheirWorld.fromWire("v2|Tehran|22:14|210|clear|27|night|1"))
        assertNull(TheirWorld.fromWire("v1||22:14|210|clear|27|night|1"))
        assertNull(TheirWorld.fromWire("v1|Tehran|22:14|2000|clear|27|night|1"))
        assertNull(TheirWorld.fromWire("v1|Tehran|22:14|210|bogus|27|night|1"))
        assertNull(TheirWorld.fromWire("x".repeat(TheirWorld.MAX_WIRE_CHARS + 1)))
    }

    @Test
    fun enumWireValues_roundTrip() {
        for (weather in WeatherCondition.entries) {
            assertEquals(weather, WeatherCondition.fromWire(weather.wireValue))
        }
        for (state in DayNightState.entries) {
            assertEquals(state, DayNightState.fromWire(state.wireValue))
        }
    }

    @Test(expected = IllegalArgumentException::class)
    fun constructor_rejectsControlCharsInLocation() {
        TheirWorld(
            locationLabel = "Bad\nCity",
            localTimeLabel = "22:14",
            utcOffsetMinutes = 210,
            weather = WeatherCondition.UNKNOWN,
            temperatureCelsius = null,
            dayNight = DayNightState.NIGHT,
            observedAtEpochSeconds = 1,
        )
    }
}
