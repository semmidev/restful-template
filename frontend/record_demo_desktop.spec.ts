import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 2,
  video: 'on',
});

test('capture comprehensive desktop demo video', async ({ page }) => {
  test.setTimeout(90000); // Set high timeout for slow execution

  const assetsDir = '/Users/sammidev/Downloads/restful-template/assets';
  const artifactDir = '/Users/sammidev/.gemini/antigravity-ide/brain/fe3ff951-0aea-4214-bab2-52591d41ba33';

  // Inject macOS Window Frame overlay dynamically on every page load
  await page.addInitScript(() => {
    const style = document.createElement('style');
    style.innerHTML = `
      body {
        background-color: #0b0813 !important;
        background-image: 
          radial-gradient(at 0% 0%, hsla(253, 16%, 7%, 1) 0px, transparent 50%),
          radial-gradient(at 50% 0%, hsla(225, 39%, 25%, 0.4) 0px, transparent 50%),
          radial-gradient(at 100% 0%, hsla(339, 49%, 25%, 0.4) 0px, transparent 50%),
          radial-gradient(at 0% 100%, hsla(271, 39%, 10%, 1) 0px, transparent 50%),
          radial-gradient(at 100% 100%, hsla(250, 50%, 15%, 1) 0px, transparent 50%) !important;
        margin: 0 !important;
        padding: 0 !important;
        display: flex !important;
        align-items: center !important;
        justify-content: center !important;
        height: 100vh !important;
        overflow: hidden !important;
      }
      
      #mac-container {
        width: 1600px;
        height: 900px;
        background: hsl(var(--background, 240 10% 4%)) !important;
        border: 1px solid rgba(255, 255, 255, 0.12) !important;
        border-radius: 16px !important;
        box-shadow: 
          0 40px 100px rgba(0, 0, 0, 0.85),
          0 12px 40px rgba(0, 0, 0, 0.65),
          0 0 0 1px rgba(0, 0, 0, 0.4),
          inset 0 1px 0 rgba(255, 255, 255, 0.15) !important;
        display: flex !important;
        flex-direction: column !important;
        overflow: hidden !important;
        position: relative !important;
        animation: scaleIn 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards !important;
      }

      @keyframes scaleIn {
        from { transform: scale(0.97); opacity: 0; }
        to { transform: scale(1); opacity: 1; }
      }

      #mac-titlebar {
        height: 44px;
        background: rgba(18, 18, 24, 0.8) !important;
        backdrop-filter: blur(20px) !important;
        border-bottom: 1px solid rgba(255, 255, 255, 0.08) !important;
        display: flex !important;
        align-items: center !important;
        padding: 0 18px !important;
        position: relative !important;
        user-select: none !important;
        flex-shrink: 0 !important;
        z-index: 99999 !important;
      }

      .mac-dots {
        display: flex !important;
        gap: 8px !important;
      }

      .mac-dot {
        width: 12px !important;
        height: 12px !important;
        border-radius: 50% !important;
        position: relative !important;
      }

      .mac-close { 
        background-color: #ff5f56 !important; 
        border: 0.5px solid #e0443e !important;
      }
      .mac-minimize { 
        background-color: #ffbd2e !important; 
        border: 0.5px solid #dfa223 !important;
      }
      .mac-maximize { 
        background-color: #27c93f !important; 
        border: 0.5px solid #1aab29 !important;
      }

      .mac-title {
        position: absolute !important;
        left: 50% !important;
        transform: translateX(-50%) !important;
        font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", Roboto, sans-serif !important;
        font-size: 13px !important;
        font-weight: 500 !important;
        color: rgba(255, 255, 255, 0.75) !important;
        letter-spacing: 0.01em !important;
      }

      #mac-content {
        flex: 1 !important;
        position: relative !important;
        overflow: hidden !important;
        display: flex !important;
        flex-direction: column !important;
      }

      #root {
        height: 100% !important;
        width: 100% !important;
        display: flex !important;
        flex-direction: column !important;
        position: absolute !important;
        top: 0 !important;
        left: 0 !important;
      }
      
      .min-h-svh, .min-h-screen {
        min-height: 100% !important;
        height: 100% !important;
      }
    `;

    document.documentElement.appendChild(style);

    const observer = new MutationObserver((mutations, obs) => {
      const root = document.getElementById('root');
      if (root && !document.getElementById('mac-container')) {
        const container = document.createElement('div');
        container.id = 'mac-container';

        const titlebar = document.createElement('div');
        titlebar.id = 'mac-titlebar';
        titlebar.innerHTML = `
          <div class="mac-dots">
            <div class="mac-dot mac-close"></div>
            <div class="mac-dot mac-minimize"></div>
            <div class="mac-dot mac-maximize"></div>
          </div>
          <div class="mac-title">TodoApp — Personal Workspace</div>
        `;

        const content = document.createElement('div');
        content.id = 'mac-content';

        obs.disconnect();

        root.parentNode?.insertBefore(container, root);
        container.appendChild(titlebar);
        container.appendChild(content);
        content.appendChild(root);

        obs.observe(document.body, { childList: true, subtree: true });
      }
    });

    observer.observe(document.documentElement, { childList: true, subtree: true });
  });

  // 1. Go to Login page
  await page.goto('http://localhost:8080/login');
  await page.waitForTimeout(2000);

  // 2. Go to Register page
  await page.goto('http://localhost:8080/register');
  await page.waitForTimeout(2000);

  // 3. Register new user
  const email = `demo_user_desktop_${Date.now()}@example.com`;
  await page.fill('#email', email);
  await page.fill('#password', 'Password123');
  await page.fill('#confirmPassword', 'Password123');
  await page.click('button[type="submit"]');

  // 4. Wait for redirect to Dashboard
  await page.waitForURL('http://localhost:8080/');
  await page.waitForTimeout(2000);

  // 5. Go to Tasks page to add tasks
  await page.goto('http://localhost:8080/todos');
  await page.waitForTimeout(2000);

  // Helper function to create task
  const createTask = async (title: string, desc: string) => {
    await page.click('button:has-text("New Task")');
    await page.waitForSelector('input[placeholder="Task title..."]');
    await page.fill('input[placeholder="Task title..."]', title);
    await page.fill('textarea[placeholder="Add task details & description..."]', desc);
    await page.click('button:has-text("Create Task")');
    await page.waitForTimeout(1500);
  };

  // Add multiple tasks
  await createTask('Implement Huma v2 API Routes', 'Set up clean hexagonal delivery handlers.');
  await createTask('Polish Frontend Layout Aesthetics', 'Use HSL CSS variables and clean glassmorphism.');
  await createTask('Integrate Tracing Telemetry', 'Connect OpenTelemetry export pipeline.');
  await createTask('Add Redis Caching Middleware', 'Distribute query load asynchronously.');
  await createTask('Set Up Docker Compose Distroless', 'Minimize deployment image payload sizes.');
  await createTask('Refactor Modular Boundaries', 'Clean hexagonal architectural isolation.');

  // Toggle statuses
  // Start 'Polish Frontend Layout Aesthetics'
  const task2Row = page.locator('div.group:has-text("Polish Frontend Layout Aesthetics")').first();
  await task2Row.locator('button:has-text("Start")').click();
  await page.waitForTimeout(1500);

  // Start 'Set Up Docker Compose Distroless'
  const task5Row = page.locator('div.group:has-text("Set Up Docker Compose Distroless")').first();
  await task5Row.locator('button:has-text("Start")').click();
  await page.waitForTimeout(1500);

  // Finish 'Implement Huma v2 API Routes'
  const task1Row = page.locator('div.group:has-text("Implement Huma v2 API Routes")').first();
  await task1Row.locator('button:has-text("Finish")').click();
  await page.waitForTimeout(1500);

  // Finish 'Integrate Tracing Telemetry'
  const task3Row = page.locator('div.group:has-text("Integrate Tracing Telemetry")').first();
  await task3Row.locator('button:has-text("Finish")').click();
  await page.waitForTimeout(2000);

  // Scroll task list down and up
  const mainContent = page.locator('main').first();
  await mainContent.evaluate(el => el.scrollBy(0, 300));
  await page.waitForTimeout(1500);
  await mainContent.evaluate(el => el.scrollBy(0, -300));
  await page.waitForTimeout(1500);

  // 6. Go to Matrix page
  await page.goto('http://localhost:8080/matrix');
  await page.waitForTimeout(3000);

  // Perform Drag and Drop
  const card = page.locator('div[draggable]:has-text("Add Redis Caching Middleware")').first();
  const doFirstQuadrant = page.locator('div:has-text("Do First")').last();
  await card.dragTo(doFirstQuadrant);
  await page.waitForTimeout(3000);

  // 7. Go back to Dashboard page
  await page.goto('http://localhost:8080/');
  await page.waitForTimeout(4000);

  // Scroll dashboard to show Recharts area chart and bar chart
  await mainContent.evaluate(el => el.scrollBy(0, 400));
  await page.waitForTimeout(4000);

  // Take thumbnail screenshot of the final dashboard
  const thumbnailPath = path.join(assetsDir, 'desktop-demo-thumbnail.png');
  await page.screenshot({ path: thumbnailPath });
  fs.copyFileSync(thumbnailPath, path.join(artifactDir, 'desktop_demo_thumbnail.png'));
  console.log('Thumbnail screenshot successfully captured!');

  // 8. Close page and copy video
  const video = page.video();
  await page.close();

  if (video) {
    const videoPath = await video.path();
    if (fs.existsSync(videoPath)) {
      // Copy to assets directory
      fs.copyFileSync(videoPath, path.join(assetsDir, 'desktop-demo.webm'));
      // Copy to conversation artifacts directory
      fs.copyFileSync(videoPath, path.join(artifactDir, 'desktop_demo.webm'));
      console.log('Video successfully recorded and saved!');
    }
  }
});
