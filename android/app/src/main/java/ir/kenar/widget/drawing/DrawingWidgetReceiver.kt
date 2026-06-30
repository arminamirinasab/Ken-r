package ir.kenar.widget.drawing

import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver

/**
 * Manifest-registered receiver that hosts [DrawingWidget].
 *
 * Drawing capture belongs in the app UI; this widget stays a passive renderer
 * for the partner's latest synced sketch.
 */
class DrawingWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = DrawingWidget()
}
