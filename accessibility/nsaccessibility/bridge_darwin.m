// +build darwin

#import <Cocoa/Cocoa.h>
#import <AppKit/AppKit.h>

// Accessibility element wrapper
@interface FluffyAccessibilityElement : NSAccessibilityElement
@property (nonatomic, copy) NSString* accessibilityLabel;
@property (nonatomic, copy) NSString* accessibilityRole;
@property (nonatomic, copy) NSString* accessibilityValue;
@property (nonatomic, assign) BOOL accessibilityEnabled;
@property (nonatomic, assign) NSRect accessibilityFrame;
@property (nonatomic, weak) id accessibilityParent;
@property (nonatomic, strong) NSArray* accessibilityChildren;
@end

@implementation FluffyAccessibilityElement
@end

// Global state
static BOOL g_initialized = NO;
static FluffyAccessibilityElement* g_rootElement = nil;

// Initialize accessibility system
void nsaccessibility_init(void) {
    if (g_initialized) {
        return;
    }

    @autoreleasepool {
        // Enable accessibility
        NSApplicationLoad();

        // Create root accessibility element
        g_rootElement = [[FluffyAccessibilityElement alloc] init];
        g_rootElement.accessibilityRole = NSAccessibilityApplicationRole;
        g_rootElement.accessibilityLabel = [[NSProcessInfo processInfo] processName];
        g_rootElement.accessibilityEnabled = YES;

        g_initialized = YES;
    }
}

// Cleanup accessibility system
void nsaccessibility_cleanup(void) {
    if (!g_initialized) {
        return;
    }

    @autoreleasepool {
        g_rootElement = nil;
        g_initialized = NO;
    }
}

// Announce text to VoiceOver
void nsaccessibility_announce(const char* message, int priority) {
    if (!g_initialized || message == NULL) {
        return;
    }

    @autoreleasepool {
        NSString* text = [NSString stringWithUTF8String:message];
        if (text.length == 0) {
            return;
        }

        // Post announcement notification
        NSDictionary* userInfo = @{
            NSAccessibilityAnnouncementKey: text,
            NSAccessibilityPriorityKey: priority > 0 ?
                @(NSAccessibilityPriorityHigh) : @(NSAccessibilityPriorityMedium)
        };

        NSAccessibilityPostNotificationWithUserInfo(
            g_rootElement,
            NSAccessibilityAnnouncementRequestedNotification,
            userInfo
        );
    }
}

// Notify focus change
void nsaccessibility_notify_focus_change(const char* label, const char* role) {
    if (!g_initialized) {
        return;
    }

    @autoreleasepool {
        // Create a temporary element for the focused widget
        FluffyAccessibilityElement* element = [[FluffyAccessibilityElement alloc] init];
        if (label) {
            element.accessibilityLabel = [NSString stringWithUTF8String:label];
        }
        if (role) {
            element.accessibilityRole = [NSString stringWithUTF8String:role];
        }
        element.accessibilityEnabled = YES;
        element.accessibilityParent = g_rootElement;

        // Post focus notification
        NSAccessibilityPostNotification(element, NSAccessibilityFocusedUIElementChangedNotification);
    }
}

// Notify value change
void nsaccessibility_notify_value_change(const char* label, const char* old_value, const char* new_value) {
    if (!g_initialized) {
        return;
    }

    @autoreleasepool {
        FluffyAccessibilityElement* element = [[FluffyAccessibilityElement alloc] init];
        if (label) {
            element.accessibilityLabel = [NSString stringWithUTF8String:label];
        }
        if (new_value) {
            element.accessibilityValue = [NSString stringWithUTF8String:new_value];
        }
        element.accessibilityEnabled = YES;
        element.accessibilityParent = g_rootElement;

        // Post value change notification
        NSAccessibilityPostNotification(element, NSAccessibilityValueChangedNotification);
    }
}

// Notify state change
void nsaccessibility_notify_state_change(const char* label, const char* state, int value) {
    if (!g_initialized || state == NULL) {
        return;
    }

    @autoreleasepool {
        FluffyAccessibilityElement* element = [[FluffyAccessibilityElement alloc] init];
        if (label) {
            element.accessibilityLabel = [NSString stringWithUTF8String:label];
        }
        element.accessibilityEnabled = YES;
        element.accessibilityParent = g_rootElement;

        NSString* stateName = [NSString stringWithUTF8String:state];

        // Post appropriate notification based on state type
        if ([stateName isEqualToString:@"expanded"]) {
            if (value) {
                NSAccessibilityPostNotification(element, NSAccessibilityRowExpandedNotification);
            } else {
                NSAccessibilityPostNotification(element, NSAccessibilityRowCollapsedNotification);
            }
        } else if ([stateName isEqualToString:@"selected"]) {
            NSAccessibilityPostNotification(element, NSAccessibilitySelectedChildrenChangedNotification);
        } else {
            // Generic state change - use layout changed
            NSAccessibilityPostNotification(element, NSAccessibilityLayoutChangedNotification);
        }
    }
}

// Check if VoiceOver is running
int nsaccessibility_is_voiceover_running(void) {
    @autoreleasepool {
        return NSWorkspace.sharedWorkspace.isVoiceOverEnabled ? 1 : 0;
    }
}
