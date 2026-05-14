import { Page } from '@playwright/test';

/**
 * Helper class for all prompt/dialog interactions.
 * Provides stable API decoupled from UI selectors.
 */
export class PromptHelper {
  constructor(private page: Page) {}

  /**
   * Wait for any prompt or overlay to appear on screen.
   */
  async waitForPrompt(timeout = 5000): Promise<void> {
    await this.page.waitForSelector(
      '[data-testid="prompt-dialog"], ' +
      '[data-testid="decision-overlay"], ' +
      '[data-testid="skill-branch-overlay"], ' +
      '[data-testid="skill-select-panel"]',
      {
        timeout,
        state: 'visible',
      }
    );
  }

  /**
   * Click confirm button on a confirmation prompt.
   */
  async confirm(): Promise<void> {
    await this.page.click('[data-testid="prompt-confirm-btn"]');
    await this.waitForDialogClose();
  }

  /**
   * Click cancel button on a confirmation prompt.
   */
  async cancel(): Promise<void> {
    await this.page.click('[data-testid="prompt-cancel-btn"]');
    await this.waitForDialogClose();
  }

  /**
   * Select a branch option by index.
   * Works for skill-branch-overlay and decision-overlay.
   */
  async selectBranch(branchIndex: number): Promise<void> {
    await this.page.click(`[data-testid="branch-option-${branchIndex}"]`);
    await this.waitForDialogClose();
  }

  /**
   * Select a numeric value from overlay buttons.
   * Used for X/Y value selection (like plague_death_touch_x).
   */
  async selectNumericValue(value: number): Promise<void> {
    // The numeric button displays the value directly
    await this.page.click(`[data-testid="numeric-option-${value}"]`);
    await this.waitForDialogClose();
  }

  /**
   * Input a number value (legacy - for numeric input fields).
   */
  async inputNumber(value: number): Promise<void> {
    // Try numeric button first
    const numericBtn = await this.page.$(`[data-testid="numeric-option-${value}"]`);
    if (numericBtn) {
      await numericBtn.click();
      await this.waitForDialogClose();
      return;
    }
    // Fallback to input field
    await this.page.fill('[data-testid="numeric-input"]', String(value));
    await this.page.click('[data-testid="numeric-confirm-btn"]');
    await this.waitForDialogClose();
  }

  /**
   * Select multiple cards by indices.
   * For card selection prompts (PromptChooseCards).
   * Cards are selected from hand area or inline options.
   */
  async selectCards(cardIndices: number[]): Promise<void> {
    for (const index of cardIndices) {
      // Try inline option first (card-picker style)
      const inlineOption = await this.page.$(`[data-testid="card-${index}"]`);
      if (inlineOption) {
        await inlineOption.click();
        continue;
      }
      // Fallback to hand card selection
      await this.page.click(`[data-testid="hand-card-${index}"]`);
    }

    // Try to confirm with explicit confirm button
    const confirmBtn = await this.page.$('[data-testid="card-picker-confirm-btn"]');
    if (confirmBtn) {
      await confirmBtn.click();
    }
    await this.waitForDialogClose();
  }

  /**
   * Select single card by index.
   */
  async selectCard(cardIndex: number): Promise<void> {
    await this.selectCards([cardIndex]);
  }

  /**
   * Select a single target player.
   * Uses data-player-anchor attribute on player area.
   */
  async selectTarget(playerId: string): Promise<void> {
    await this.page.click(`[data-player-anchor="${playerId}"]`);
    await this.waitForDialogClose();
  }

  /**
   * Select multiple target players.
   */
  async selectTargets(playerIds: string[]): Promise<void> {
    for (const id of playerIds) {
      await this.page.click(`[data-player-anchor="${id}"]`);
    }
    await this.waitForDialogClose();
  }

  /**
   * Get the current prompt type (data-testid).
   */
  async getPromptType(): Promise<string> {
    const dialog = await this.page.waitForSelector(
      '[data-testid="prompt-dialog"], ' +
      '[data-testid="decision-overlay"], ' +
      '[data-testid="skill-branch-overlay"], ' +
      '[data-testid="skill-select-panel"]'
    );
    const testId = await dialog.getAttribute('data-testid');
    return testId || '';
  }

  /**
   * Get the prompt message text.
   */
  async getPromptMessage(): Promise<string> {
    const messageEl = await this.page.waitForSelector('[data-testid="prompt-message"], .overlay-panel-header h2');
    return await messageEl.textContent() || '';
  }

  /**
   * Wait for dialog to close.
   */
  private async waitForDialogClose(timeout = 2000): Promise<void> {
    await this.page.waitForSelector(
      '[data-testid="prompt-dialog"], ' +
      '[data-testid="decision-overlay"], ' +
      '[data-testid="skill-branch-overlay"], ' +
      '[data-testid="skill-select-panel"]',
      {
        timeout,
        state: 'hidden',
      }
    ).catch(() => {});
  }
}