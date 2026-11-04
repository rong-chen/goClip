//go:build darwin

#include <stdint.h>
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

extern void goTrayToggleWindow();
extern void goTrayShowWindow();
extern void goTrayOpenSettings();
extern void goTrayQuitApp();

static void bringMainWindowToFront(void) {
    for (NSWindow *window in [NSApp windows]) {
        if ([window canBecomeKeyWindow]) {
            [window makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:YES];
            break;
        }
    }
    goTrayShowWindow();
}

static BOOL custom_reopen(id self, SEL _cmd, NSApplication *sender, BOOL hasVisibleWindows) {
    bringMainWindowToFront();
    return YES;
}

static void hookAppDelegateReopen(void) {
    id appDelegate = [NSApp delegate];
    if (appDelegate != nil) {
        Class delClass = [appDelegate class];
        Method origMethod = class_getInstanceMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:));
        if (origMethod != NULL) {
            class_replaceMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:), (IMP)custom_reopen, "c@:@c");
        } else {
            class_addMethod(delClass, @selector(applicationShouldHandleReopen:hasVisibleWindows:), (IMP)custom_reopen, "c@:@c");
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
@end

@implementation GoClipTrayController

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

        hookAppDelegateReopen();

        [[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationDidBecomeActiveNotification 
                                                          object:nil 
                                                           queue:[NSOperationQueue mainQueue] 
                                                      usingBlock:^(NSNotification *note) {
            bringMainWindowToFront();
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
        goTrayToggleWindow();
    }
}

- (void)onMenuToggle:(id)sender {
    goTrayToggleWindow();
}

- (void)onMenuSettings:(id)sender {
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
