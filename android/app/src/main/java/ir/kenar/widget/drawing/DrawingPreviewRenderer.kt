package ir.kenar.widget.drawing

import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Paint
import ir.kenar.domain.widget.SharedDrawing

/**
 * Renders a tiny logical drawing into a bitmap that Glance can show.
 *
 * The bitmap is intentionally small to respect widget IPC limits. The full
 * drawing payload remains vector-like and compact; this renderer is only the
 * local widget preview.
 */
object DrawingPreviewRenderer {
    private const val BITMAP_SIZE = 256
    private const val BACKGROUND = 0xFFFFF7F9.toInt()

    fun render(drawing: SharedDrawing): Bitmap {
        val bitmap = Bitmap.createBitmap(BITMAP_SIZE, BITMAP_SIZE, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        canvas.drawColor(BACKGROUND)

        val scaleX = BITMAP_SIZE / drawing.canvasWidth.toFloat()
        val scaleY = BITMAP_SIZE / drawing.canvasHeight.toFloat()
        val paint = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            style = Paint.Style.STROKE
            strokeCap = Paint.Cap.ROUND
            strokeJoin = Paint.Join.ROUND
        }

        drawing.strokes.forEach { stroke ->
            paint.color = stroke.colorArgb
            paint.strokeWidth = stroke.widthPx * ((scaleX + scaleY) / 2f)

            val points = stroke.points
            when (points.size) {
                0 -> Unit
                1 -> {
                    paint.style = Paint.Style.FILL
                    canvas.drawCircle(points[0].x * scaleX, points[0].y * scaleY, paint.strokeWidth / 2f, paint)
                    paint.style = Paint.Style.STROKE
                }
                else -> points.zipWithNext().forEach { (from, to) ->
                    canvas.drawLine(from.x * scaleX, from.y * scaleY, to.x * scaleX, to.y * scaleY, paint)
                }
            }
        }

        return bitmap
    }
}
