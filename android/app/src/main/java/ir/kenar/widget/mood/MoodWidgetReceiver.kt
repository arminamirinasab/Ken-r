package ir.kenar.widget.mood

import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver

/**
 * Manifest-registered receiver that hosts [MoodWidget]. The sync layer updates
 * the widget by writing its Glance state and calling update(); this receiver
 * itself does no polling (updatePeriodMillis = 0).
 */
class MoodWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = MoodWidget()
}
