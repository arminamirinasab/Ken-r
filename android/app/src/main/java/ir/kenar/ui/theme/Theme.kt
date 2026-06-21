package ir.kenar.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

private val LightScheme = lightColorScheme(
    primary = KenarColors.LightPrimary,
    onPrimary = KenarColors.LightOnPrimary,
    background = KenarColors.LightBackground,
    surface = KenarColors.LightSurface,
    onSurface = KenarColors.LightOnSurface,
)

private val DarkScheme = darkColorScheme(
    primary = KenarColors.DarkPrimary,
    onPrimary = KenarColors.DarkOnPrimary,
    background = KenarColors.DarkBackground,
    surface = KenarColors.DarkSurface,
    onSurface = KenarColors.DarkOnSurface,
)

/**
 * Root theme. RTL/LTR is driven by the active locale via the platform layout
 * direction, so no manual mirroring is needed in Compose layouts.
 */
@Composable
fun KenarTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    MaterialTheme(
        colorScheme = if (darkTheme) DarkScheme else LightScheme,
        typography = Typography(),
        content = content,
    )
}
