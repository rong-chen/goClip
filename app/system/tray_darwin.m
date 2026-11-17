//go:build darwin

#include <stdint.h>
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

extern void goTrayToggleWindow();
extern void goTrayShowWindow();
extern void goTrayHideWindow();
extern void goTrayOpenSettings();
extern void goTrayQuitApp();

static NSWindow *findMainWindow(void) {
    id delegate = [NSApp delegate];
    if (delegate != nil && [delegate respondsToSelector:@selector(mainWindow)]) {
        NSWindow *w = [delegate performSelector:@selector(mainWindow)];
        if (w != nil) {
            return w;
        }
    }
    for (NSWindow *window in [NSApp windows]) {
        if ([window isKindOfClass:NSClassFromString(@"NSStatusBarWindow")]) {
            continue;
        }
        if ([window canBecomeKeyWindow]) {
            return window;
        }
    }
    if ([NSApp mainWindow] != nil) {
        return [NSApp mainWindow];
    }
    if ([[NSApp windows] count] > 0) {
        return [[NSApp windows] firstObject];
    }
    return nil;
}

int isMacWindowFrontAndActive(void) {
    if ([NSApp isHidden]) {
        return 0;
    }
    if (![NSApp isActive]) {
        return 0;
    }
    NSWindow *mainWindow = findMainWindow();
    if (mainWindow == nil) {
        return 0;
    }
    if (![mainWindow isVisible] || [mainWindow isMiniaturized]) {
        return 0;
    }
    return 1;
}

static void bringMainWindowToFront(void) {
    if ([NSApp isHidden]) {
        [NSApp unhide:nil];
    }
    [NSApp activateIgnoringOtherApps:YES];

    NSWindow *mainWindow = findMainWindow();
    if (mainWindow != nil) {
        if ([mainWindow isMiniaturized]) {
            [mainWindow deminiaturize:nil];
        }
        [mainWindow makeKeyAndOrderFront:nil];
        [mainWindow orderFrontRegardless];
    }
    goTrayShowWindow();
}

static void hideMainWindow(void) {
    NSWindow *mainWindow = findMainWindow();
    if (mainWindow != nil) {
        [mainWindow orderOut:nil];
    }
    [NSApp hide:nil];
    goTrayHideWindow();
}

static BOOL custom_reopen(id self, SEL _cmd, NSApplication *sender, BOOL hasVisibleWindows) {
    bringMainWindowToFront();
    return YES;
}

static void hookAppDelegateReopen(void) {
    Class delClass = objc_getClass("AppDelegate");
    if (delClass == nil && [NSApp delegate] != nil) {
        delClass = [[NSApp delegate] class];
    }
    if (delClass != nil) {
        Method origMethod = class_getInstanceMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:));
        if (origMethod != NULL) {
            class_replaceMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:), (IMP)custom_reopen, "B@:@B");
        } else {
            class_addMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:), (IMP)custom_reopen, "B@:@B");
        }
        if ([NSApp delegate] != nil) {
            [NSApp setDelegate:[NSApp delegate]];
        }
    }
}

@interface GoClipTrayController : NSObject
@property (strong) NSStatusItem *statusItem;
@property (strong) NSMenu *trayMenu;
- (void)setupTrayWithBytes:(const void *)bytes length:(int)len;
- (void)onButtonClick:(id)sender;
- (void)onMenuToggle:(id)sender;
- (void)onMenuSettings:(id)sender;
- (void)onMenuQuit:(id)sender;
- (void)handleReopenEvent:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
@end

@implementation GoClipTrayController

- (void)handleReopenEvent:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
    bringMainWindowToFront();
}

- (void)setupTrayWithBytes:(const void *)bytes length:(int)len {
    dispatch_async(dispatch_get_main_queue(), ^{
        self.statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];

        NSData *data = [NSData dataWithBytes:bytes length:len];
        NSImage *image = [[NSImage alloc] initWithData:data];
        [image setTemplate:YES];
        [image setSize:NSMakeSize(18, 18)];

        NSStatusBarButton *button = self.statusItem.button;
        if (button != nil) {
            button.image = image;
            button.toolTip = @"GoClip";
            button.target = self;
            button.action = @selector(onButtonClick:);
            [button sendActionOn:(NSEventMaskLeftMouseUp | NSEventMaskRightMouseUp)];
        }

        self.trayMenu = [[NSMenu alloc] initWithTitle:@"GoClip"];

        NSMenuItem *toggleItem = [[NSMenuItem alloc] initWithTitle:@"显示或隐藏窗口" action:@selector(onMenuToggle:) keyEquivalent:@""];
        toggleItem.target = self;
        [self.trayMenu addItem:toggleItem];

        NSMenuItem *settingsItem = [[NSMenuItem alloc] initWithTitle:@"偏好设置" action:@selector(onMenuSettings:) keyEquivalent:@","];
        settingsItem.target = self;
        [self.trayMenu addItem:settingsItem];

        [self.trayMenu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"退出 GoClip" action:@selector(onMenuQuit:) keyEquivalent:@"q"];
        quitItem.target = self;
        [self.trayMenu addItem:quitItem];

        [[NSAppleEventManager sharedAppleEventManager] setEventHandler:self
                                                           andSelector:@selector(handleReopenEvent:withReplyEvent:)
                                                         forEventClass:kCoreEventClass
                                                            andEventID:kAEReopenApplication];

        hookAppDelegateReopen();

        [[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationDidBecomeActiveNotification 
                                                          object:nil 
                                                           queue:[NSOperationQueue mainQueue] 
                                                      usingBlock:^(NSNotification *note) {
            bringMainWindowToFront();
        }];

        [[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationDidHideNotification 
                                                          object:nil 
                                                           queue:[NSOperationQueue mainQueue] 
                                                      usingBlock:^(NSNotification *note) {
            goTrayHideWindow();
        }];

        [[NSNotificationCenter defaultCenter] addObserverForName:NSWindowDidMiniaturizeNotification 
                                                          object:nil 
                                                           queue:[NSOperationQueue mainQueue] 
                                                      usingBlock:^(NSNotification *note) {
            goTrayHideWindow();
        }];

        [[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationDidFinishLaunchingNotification 
                                                          object:nil 
                                                           queue:[NSOperationQueue mainQueue] 
                                                      usingBlock:^(NSNotification *note) {
            hookAppDelegateReopen();
        }];
    });
}

- (void)onButtonClick:(id)sender {
    NSEvent *event = [NSApp currentEvent];
    BOOL isRightClick = (event.type == NSEventTypeRightMouseUp || 
                         event.type == NSEventTypeRightMouseDown || 
                         (event.modifierFlags & NSEventModifierFlagControl));

    if (isRightClick) {
        [self.statusItem setMenu:self.trayMenu];
        [self.statusItem.button performClick:nil];
        [self.statusItem setMenu:nil];
    } else {
        if (isMacWindowFrontAndActive()) {
            hideMainWindow();
        } else {
            bringMainWindowToFront();
        }
    }
}

- (void)onMenuToggle:(id)sender {
    if (isMacWindowFrontAndActive()) {
        hideMainWindow();
    } else {
        bringMainWindowToFront();
    }
}

- (void)onMenuSettings:(id)sender {
    bringMainWindowToFront();
    goTrayOpenSettings();
}

- (void)onMenuQuit:(id)sender {
    goTrayQuitApp();
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(100 * NSEC_PER_MSEC)), dispatch_get_main_queue(), ^{
        [NSApp terminate:nil];
    });
}

@end

static GoClipTrayController *globalTray = nil;

void initMacTray(const void *bytes, int len) {
    if (globalTray == nil) {
        globalTray = [[GoClipTrayController alloc] init];
        [globalTray setupTrayWithBytes:bytes length:len];
    }
}
