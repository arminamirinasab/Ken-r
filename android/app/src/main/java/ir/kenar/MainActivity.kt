package ir.kenar

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import dagger.hilt.android.AndroidEntryPoint
import ir.kenar.core.locale.LocaleManager
import ir.kenar.ui.theme.KenarTheme

/**
 * Single-activity host. Real navigation (pairing → shared space → settings) is
 * added as those screens land. For now it shows the wordmark, tagline, and the
 * in-app language switch so the bilingual pipeline is verifiable from day one.
 */
@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            KenarTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    LandingScreen()
                }
            }
        }
    }
}

@Composable
private fun LandingScreen() {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = stringResource(R.string.app_name),
            style = MaterialTheme.typography.displaySmall,
        )
        Spacer(Modifier.height(8.dp))
        Text(
            text = stringResource(R.string.app_tagline),
            style = MaterialTheme.typography.bodyLarge,
        )
        Spacer(Modifier.height(32.dp))
        Button(onClick = { /* TODO: navigate to pairing */ }) {
            Text(stringResource(R.string.pair_create_invite))
        }
        Spacer(Modifier.height(16.dp))
        TextButton(onClick = { LocaleManager.set(LocaleManager.Language.PERSIAN) }) {
            Text(stringResource(R.string.settings_language_fa))
        }
        TextButton(onClick = { LocaleManager.set(LocaleManager.Language.ENGLISH) }) {
            Text(stringResource(R.string.settings_language_en))
        }
    }
}

@Preview(showBackground = true)
@Composable
private fun LandingPreview() {
    KenarTheme { LandingScreen() }
}
