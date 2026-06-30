package ir.kenar.widget.world

import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver

/**
 * Manifest-registered receiver that hosts [TheirWorldWidget].
 *
 * The receiver does no weather/location work; the app sync layer prepares the
 * snapshot and refreshes Glance state.
 */
class TheirWorldWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = TheirWorldWidget()
}
