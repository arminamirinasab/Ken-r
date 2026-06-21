package ir.kenar.domain.widget

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class MoodTest {

    @Test
    fun fromWire_roundTripsEveryMood() {
        for (mood in Mood.entries) {
            assertEquals(mood, Mood.fromWire(mood.wireValue))
        }
    }

    @Test
    fun fromWire_returnsNullForUnknownOrNull() {
        assertNull(Mood.fromWire("unknown"))
        assertNull(Mood.fromWire(null))
    }

    @Test
    fun wireValues_areUnique() {
        val values = Mood.entries.map { it.wireValue }
        assertEquals(values.size, values.toSet().size)
    }
}
