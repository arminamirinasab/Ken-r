package ir.kenar.widget.lovetap

import android.content.Context
import androidx.datastore.preferences.core.toMutablePreferences
import androidx.glance.GlanceId
import androidx.glance.action.ActionParameters
import androidx.glance.appwidget.action.ActionCallback
import androidx.glance.appwidget.state.updateAppWidgetState
import androidx.glance.state.PreferencesGlanceStateDefinition
import ir.kenar.domain.widget.LoveTap

/**
 * Queues an outgoing Love Tap from the widget surface.
 *
 * The networking sync layer will own upload/retry semantics; this callback only
 * records the user's tap in Glance state so the widget remains responsive and
 * battery-friendly while staying aligned with the widget-first architecture.
 */
class LoveTapActionCallback : ActionCallback {
    override suspend fun onAction(
        context: Context,
        glanceId: GlanceId,
        parameters: ActionParameters,
    ) {
        val wireValue = parameters[LoveTapWidget.TAP_ACTION_PARAM] ?: return
        val tap = LoveTap.fromWire(wireValue) ?: return

        updateAppWidgetState(context, PreferencesGlanceStateDefinition, glanceId) { prefs ->
            prefs.toMutablePreferences().apply {
                this[LoveTapWidget.OUTGOING_TAP_KEY] = tap.wireValue
                this[LoveTapWidget.OUTGOING_TAP_CREATED_AT_KEY] = System.currentTimeMillis().toString()
                this[LoveTapWidget.LAST_SENT_TAP_KEY] = tap.wireValue
            }
        }
        LoveTapWidget().update(context, glanceId)
    }
}
