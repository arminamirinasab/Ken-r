package ir.kenar.domain.widget

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class SharedDrawingTest {

    @Test
    fun encodeWire_roundTripsDrawing() {
        val drawing = SharedDrawing(
            strokes = listOf(
                DrawingStroke(
                    colorArgb = 0xFFE04F7A.toInt(),
                    widthPx = 3,
                    points = listOf(DrawingPoint(4, 5), DrawingPoint(12, 18), DrawingPoint(32, 20)),
                ),
                DrawingStroke(
                    colorArgb = 0xFF6D5961.toInt(),
                    widthPx = 2,
                    points = listOf(DrawingPoint(20, 30)),
                ),
            ),
        )

        assertEquals(drawing, SharedDrawing.fromWire(drawing.encodeWire()))
    }

    @Test
    fun fromWire_returnsNullForUnknownMalformedOrOversizedPayload() {
        assertNull(SharedDrawing.fromWire(null))
        assertNull(SharedDrawing.fromWire(""))
        assertNull(SharedDrawing.fromWire("v2|64|64|"))
        assertNull(SharedDrawing.fromWire("v1|0|64|"))
        assertNull(SharedDrawing.fromWire("v1|64|64|not-a-color,2,1,1"))
        assertNull(SharedDrawing.fromWire("x".repeat(SharedDrawing.MAX_WIRE_CHARS + 1)))
    }

    @Test
    fun fromNormalized_clampsCoordinatesToCanvas() {
        assertEquals(DrawingPoint(0, 64), DrawingPoint.fromNormalized(-1f, 2f))
        assertEquals(DrawingPoint(32, 16), DrawingPoint.fromNormalized(0.5f, 0.25f))
    }

    @Test
    fun emptyDrawing_roundTripsAndReportsEmpty() {
        val drawing = SharedDrawing()

        val decoded = SharedDrawing.fromWire(drawing.encodeWire())

        assertEquals(drawing, decoded)
        assertTrue(decoded?.isEmpty() == true)
    }
}
