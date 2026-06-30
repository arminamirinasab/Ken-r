package ir.kenar.widget.countdown

import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver

/**
 * Manifest-registered receiver that hosts [CountdownWidget].
 *
 * Date editing and daily snapshot refresh belong to the app sync layer; this
 * receiver only exposes the passive home-screen widget.
 */
class CountdownWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = CountdownWidget()
}
