package ir.kenar.domain.widget

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class LoveTapTest {

    @Test
    fun fromWire_roundTripsEveryLoveTap() {
        for (tap in LoveTap.entries) {
            assertEquals(tap, LoveTap.fromWire(tap.wireValue))
        }
    }

    @Test
    fun fromWire_returnsNullForUnknownOrNull() {
        assertNull(LoveTap.fromWire("unknown"))
        assertNull(LoveTap.fromWire(null))
    }

    @Test
    fun wireValues_areUnique() {
        val values = LoveTap.entries.map { it.wireValue }
        assertEquals(values.size, values.toSet().size)
    }

    @Test
    fun primaryActions_excludesTapBack() {
        assertFalse(LoveTap.primaryActions.contains(LoveTap.CAUGHT_IT))
        assertTrue(LoveTap.CAUGHT_IT.isTapBack)
    }
}
