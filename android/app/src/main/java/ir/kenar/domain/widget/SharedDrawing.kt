package ir.kenar.domain.widget

import kotlin.math.roundToInt

/**
 * Compact, widget-safe line drawing shared between paired devices.
 *
 * Coordinates live in a small logical canvas. The sync layer encrypts the wire
 * string before sending it through the generic `drawing` widget endpoint, so
 * the backend still treats the payload as an opaque blob.
 */
data class SharedDrawing(
    val canvasWidth: Int = DEFAULT_CANVAS_SIZE,
    val canvasHeight: Int = DEFAULT_CANVAS_SIZE,
    val strokes: List<DrawingStroke> = emptyList(),
) {
    fun isEmpty(): Boolean = strokes.none { it.points.isNotEmpty() }

    fun encodeWire(): String {
        require(canvasWidth in 1..MAX_CANVAS_SIZE) { "canvasWidth out of range" }
        require(canvasHeight in 1..MAX_CANVAS_SIZE) { "canvasHeight out of range" }
        require(strokes.size <= MAX_STROKES) { "too many strokes" }

        val encodedStrokes = strokes.joinToString(STROKE_SEPARATOR) { stroke ->
            require(stroke.widthPx in 1..MAX_STROKE_WIDTH) { "stroke width out of range" }
            require(stroke.points.size <= MAX_POINTS_PER_STROKE) { "too many points" }
            val color = Integer.toHexString(stroke.colorArgb).padStart(8, '0')
            val points = stroke.points.joinToString(POINT_SEPARATOR) { point ->
                require(point.x in 0..canvasWidth) { "point x out of range" }
                require(point.y in 0..canvasHeight) { "point y out of range" }
                "${point.x},${point.y}"
            }
            "$color,${stroke.widthPx},$points"
        }

        val wire = listOf(WIRE_VERSION, canvasWidth.toString(), canvasHeight.toString(), encodedStrokes)
            .joinToString(SECTION_SEPARATOR)
        require(wire.length <= MAX_WIRE_CHARS) { "drawing payload too large" }
        return wire
    }

    companion object {
        const val DEFAULT_CANVAS_SIZE = 64
        const val MAX_CANVAS_SIZE = 256
        const val MAX_STROKES = 24
        const val MAX_POINTS_PER_STROKE = 128
        const val MAX_STROKE_WIDTH = 16
        const val MAX_WIRE_CHARS = 8 * 1024

        private const val WIRE_VERSION = "v1"
        private const val SECTION_SEPARATOR = "|"
        private const val STROKE_SEPARATOR = "~"
        private const val POINT_SEPARATOR = ";"

        fun fromWire(value: String?): SharedDrawing? {
            if (value.isNullOrBlank() || value.length > MAX_WIRE_CHARS) return null

            return runCatching {
                val sections = value.split(SECTION_SEPARATOR, limit = 4)
                if (sections.size != 4 || sections[0] != WIRE_VERSION) return null

                val width = sections[1].toIntOrNull() ?: return null
                val height = sections[2].toIntOrNull() ?: return null
                if (width !in 1..MAX_CANVAS_SIZE || height !in 1..MAX_CANVAS_SIZE) return null

                val strokes = parseStrokes(sections[3], width, height) ?: return null
                if (strokes.size > MAX_STROKES) return null

                SharedDrawing(canvasWidth = width, canvasHeight = height, strokes = strokes)
            }.getOrNull()
        }

        private fun parseStrokes(value: String, canvasWidth: Int, canvasHeight: Int): List<DrawingStroke>? {
            if (value.isBlank()) return emptyList()

            val rawStrokes = value.split(STROKE_SEPARATOR)
            if (rawStrokes.size > MAX_STROKES) return null

            return rawStrokes.map { parseStroke(it, canvasWidth, canvasHeight) ?: return null }
        }

        private fun parseStroke(value: String, canvasWidth: Int, canvasHeight: Int): DrawingStroke? {
            val parts = value.split(",", limit = 3)
            if (parts.size != 3) return null

            val color = parts[0].toLongOrNull(radix = 16)?.toInt() ?: return null
            val width = parts[1].toIntOrNull() ?: return null
            if (width !in 1..MAX_STROKE_WIDTH) return null

            val points = parsePoints(parts[2], canvasWidth, canvasHeight) ?: return null

            return DrawingStroke(colorArgb = color, widthPx = width, points = points)
        }

        private fun parsePoints(value: String, canvasWidth: Int, canvasHeight: Int): List<DrawingPoint>? {
            if (value.isBlank()) return emptyList()

            val rawPoints = value.split(POINT_SEPARATOR)
            if (rawPoints.size > MAX_POINTS_PER_STROKE) return null

            return rawPoints.map { parsePoint(it, canvasWidth, canvasHeight) ?: return null }
        }

        private fun parsePoint(value: String, canvasWidth: Int, canvasHeight: Int): DrawingPoint? {
            val parts = value.split(",", limit = 2)
            if (parts.size != 2) return null

            val x = parts[0].toIntOrNull() ?: return null
            val y = parts[1].toIntOrNull() ?: return null
            if (x !in 0..canvasWidth || y !in 0..canvasHeight) return null

            return DrawingPoint(x = x, y = y)
        }
    }
}

data class DrawingStroke(
    val colorArgb: Int,
    val widthPx: Int,
    val points: List<DrawingPoint>,
)

data class DrawingPoint(val x: Int, val y: Int) {
    companion object {
        fun fromNormalized(x: Float, y: Float, canvasSize: Int = SharedDrawing.DEFAULT_CANVAS_SIZE): DrawingPoint {
            val clampedX = x.coerceIn(0f, 1f)
            val clampedY = y.coerceIn(0f, 1f)
            return DrawingPoint(
                x = (clampedX * canvasSize).roundToInt(),
                y = (clampedY * canvasSize).roundToInt(),
            )
        }
    }
}
