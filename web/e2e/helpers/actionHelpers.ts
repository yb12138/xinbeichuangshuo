import { Page } from '@playwright/test';

/**
 * Helper class for action hub interactions.
 */
export class ActionHelper {
  constructor(private page: Page) {}

  /**
   * Click a main action button from the action hub.
   */
  async clickAction(action: 'magic' | 'attack' | 'pass'): Promise<void> {
    await this.page.click(`[data-testid="action-${action}"]`);
  }

  /**
   * Open the skill selection panel.
   */
  async openSkillPanel(): Promise<void> {
    await this.clickAction('magic');
    await this.page.click('[data-testid="action-skill"]');
    await this.page.waitForSelector('[data-testid="skill-select-panel"]');
  }

  /**
   * Select a skill from the skill selection panel by skill ID.
   * After selecting, the skill execution flow begins.
   */
  async selectSkill(skillId: string): Promise<void> {
    // Wait for skill panel to be visible
    await this.page.waitForSelector('[data-testid="skill-select-panel"]');

    // Click the specific skill button
    await this.page.click(`[data-testid="skill-${skillId}"]`);

    // Panel should close after selection
    await this.page.waitForSelector('[data-testid="skill-select-panel"]', { state: 'hidden' });
  }

  /**
   * Cancel skill selection.
   */
  async cancelSkillSelection(): Promise<void> {
    await this.page.click('[data-testid="skill-cancel-btn"]');
  }

  /**
   * End current turn (pass action).
   */
  async endTurn(): Promise<void> {
    await this.clickAction('pass');
  }

  /**
   * Play a card from hand by index.
   * Note: This requires hand cards to have data-testid="hand-card-{index}"
   */
  async playCard(cardIndex: number): Promise<void> {
    await this.page.click(`[data-testid="hand-card-${cardIndex}"]`);
    // Confirm card play if needed
  }
}
