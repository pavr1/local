/**
 * Utility to load HTML partials into pages
 */

class PartialLoader {
    static async loadPartial(partialPath, targetElementId) {
        try {
            const response = await fetch(partialPath);
            if (!response.ok) {
                throw new Error(`Failed to load partial: ${response.status}`);
            }
            
            const html = await response.text();
            const targetElement = document.getElementById(targetElementId);
            
            if (!targetElement) {
                throw new Error(`Target element with ID '${targetElementId}' not found`);
            }
            
            targetElement.innerHTML = html;
            console.log(`✅ Loaded partial: ${partialPath} into #${targetElementId}`);
            
            return true;
        } catch (error) {
            console.error(`❌ Failed to load partial ${partialPath}:`, error);
            return false;
        }
    }
    
    static async loadSystemStatus(targetElementId = 'system-status-container') {
        // Determine the correct path based on current location
        const currentPath = window.location.pathname;
        const isInSubdirectory = currentPath.includes('/inventory/') || currentPath.includes('/expense/');
        const partialPath = isInSubdirectory ? '../shared/partials/system-status.html' : 'shared/partials/system-status.html';
        
        return await this.loadPartial(partialPath, targetElementId);
    }
}

// Make available globally
window.PartialLoader = PartialLoader; 