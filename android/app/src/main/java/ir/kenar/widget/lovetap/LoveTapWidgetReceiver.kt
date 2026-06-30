package ir.kenar.widget.lovetap

import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver

/**
 * Manifest-registered receiver that hosts [LoveTapWidget].
 *
 * The widget does not poll. Incoming state is pushed by the sync layer; outgoing
 * taps are queued for the sync layer through [LoveTapActionCallback].
 */
class LoveTapWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = LoveTapWidget()
}
