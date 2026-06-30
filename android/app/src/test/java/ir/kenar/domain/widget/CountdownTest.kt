package ir.kenar.domain.widget

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class CountdownTest {

    @Test
    fun encodeWire_roundTripsCountdown() {
        val countdown = Countdown(
            title = "Anniversary",
            targetEpochDay = 21000,
            remainingDays = 12,
            asOfEpochDay = 20988,
            targetLabel = "1405/04/10",
        )

        assertEquals(countdown, Countdown.fromWire(countdown.encodeWire()))
    }

    @Test
    fun state_reflectsRemainingDays() {
        assertEquals(CountdownState.UPCOMING, countdown(3).state)
        assertEquals(CountdownState.TODAY, countdown(0).state)
        assertEquals(CountdownState.PASSED, countdown(-3).state)
    }

    @Test
    fun fromWire_returnsNullForMalformedOrOversizedPayload() {
        assertNull(Countdown.fromWire(null))
        assertNull(Countdown.fromWire(""))
        assertNull(Countdown.fromWire("v2|title|1|1|1|"))
        assertNull(Countdown.fromWire("v1||1|1|1|"))
        assertNull(Countdown.fromWire("v1|title|999999|1|1|"))
        assertNull(Countdown.fromWire("x".repeat(Countdown.MAX_WIRE_CHARS + 1)))
    }

    @Test(expected = IllegalArgumentException::class)
    fun constructor_rejectsControlCharsInTitle() {
        countdown(remainingDays = 1, title = "Trip\nsoon")
    }

    @Test(expected = IllegalArgumentException::class)
    fun constructor_rejectsTooLargeRemainingDays() {
        countdown(remainingDays = Countdown.MAX_REMAINING_DAYS + 1)
    }

    private fun countdown(remainingDays: Int, title: String = "Trip"): Countdown =
        Countdown(
            title = title,
            targetEpochDay = 20000,
            remainingDays = remainingDays,
            asOfEpochDay = 19990,
        )
}
