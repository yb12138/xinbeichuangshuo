import { Page } from '@playwright/test';

/**
 * Helper class for game state assertions.
 */
export class GameStateHelper {
  constructor(private page: Page) {}

  /**
   * Assert player's hand size.
   */
  async expectHandSize(playerId: string, size: number): Promise<void> {
    // Count hand cards using the hand-card-{index} testid pattern
    const handSection = this.page.locator(`[data-player-anchor="${playerId}"]`);
    const handCards = await handSection.locator('[data-testid^="hand-card-"]').count();
    if (handCards !== size) {
      throw new Error(`Expected hand size ${size}, got ${handCards}`);
    }
  }

  /**
   * Assert player's heal (治疗) points.
   */
  async expectHealPoints(playerId: string, count: number): Promise<void> {
    // Heal is displayed in PlayerArea component
    const healEl = this.page.locator(`[data-player-anchor="${playerId}"] .heal-value`);
    const healValue = await healEl.textContent();
    if (parseInt(healValue || '0') !== count) {
      throw new Error(`Expected heal points ${count}, got ${healValue}`);
    }
  }

  /**
   * Assert player's token count.
   */
  async expectTokenCount(playerId: string, token: string, count: number): Promise<void> {
    const tokenEl = this.page.locator(`[data-player-anchor="${playerId}"] [data-token="${token}"]`);
    const tokenCount = await tokenEl.textContent();
    if (parseInt(tokenCount || '0') !== count) {
      throw new Error(`Expected token ${token} count ${count}, got ${tokenCount}`);
    }
  }

  /**
   * Assert current player by ID.
   */
  async expectCurrentPlayer(playerId: string): Promise<void> {
    // Current player is indicated in the game state
    const turnIndicator = await this.page.locator('.turn-indicator, .current-player-marker');
    // Alternative: check player area has 'current' class
    const isCurrent = await this.page.locator(`[data-player-anchor="${playerId}"]`).evaluate(el => {
      return el.classList.contains('current-turn') || el.classList.contains('is-active');
    });
    if (!isCurrent) {
      throw new Error(`Expected ${playerId} to be current player`);
    }
  }

  /**
   * Assert player's HP.
   */
  async expectHP(playerId: string, hp: number): Promise<void> {
    const hpEl = this.page.locator(`[data-player-anchor="${playerId}"] .hp-value`);
    const hpValue = await hpEl.textContent();
    if (parseInt(hpValue || '0') !== hp) {
      throw new Error(`Expected HP ${hp}, got ${hpValue}`);
    }
  }

  /**
   * Assert player's deck size.
   */
  async expectDeckSize(playerId: string, size: number): Promise<void> {
    const deckEl = this.page.locator(`[data-player-anchor="${playerId}"] .deck-size`);
    const deckSize = await deckEl.textContent();
    if (parseInt(deckSize || '0') !== size) {
      throw new Error(`Expected deck size ${size}, got ${deckSize}`);
    }
  }

  /**
   * Wait for turn change.
   */
  async waitForTurnChange(timeout = 5000): Promise<void> {
    await this.page.waitForSelector('.turn-change-indicator', { timeout });
  }

  /**
   * Assert no prompt/overlay is currently displayed.
   */
  async expectNoPrompt(): Promise<void> {
    const prompts = await this.page.locator(
      '[data-testid="prompt-dialog"], ' +
      '[data-testid="decision-overlay"], ' +
      '[data-testid="skill-branch-overlay"], ' +
      '[data-testid="skill-select-panel"]'
    ).count();
    if (prompts > 0) {
      throw new Error('Expected no prompt, but one is displayed');
    }
  }

  /**
   * Assert a specific skill was executed (by checking logs or state).
   */
  async expectSkillExecuted(skillId: string): Promise<void> {
    // This could check game logs or state changes
    // Placeholder for now
  }

  /**
   * Get current game phase.
   */
  async getGamePhase(): Promise<string> {
    const phaseEl = await this.page.locator('.game-phase-indicator').textContent();
    return phaseEl || '';
  }
}