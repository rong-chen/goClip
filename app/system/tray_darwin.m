//go:build darwin

#include <stdint.h>
#import <Cocoa/Cocoa.h>

extern void goTrayToggleWindow();
extern void goTrayOpenSettings();
extern void goTrayQuitApp();

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
        [image setTemplate:YES]; // 设置为 Template，macOS 顶栏暗黑模式下自动转为高亮纯白
        [image setSize:NSMakeSize(18, 18)];

        NSStatusBarButton *button = self.statusItem.button;
        if (button != nil) {
            button.image = image;
            button.toolTip = @"GoClip - 本地剪贴板管理器 (点击切换)";
            button.target = self;
            button.action = @selector(onButtonClick:);
            [button sendActionOn:(NSEventMaskLeftMouseUp | NSEventMaskRightMouseUp)];
        }

        // 构建右键快捷菜单
        self.trayMenu = [[NSMenu alloc] initWithTitle:@"GoClip"];

        NSMenuItem *toggleItem = [[NSMenuItem alloc] initWithTitle:@"显示/隐藏 GoClip" action:@selector(onMenuToggle:) keyEquivalent:@""];
        toggleItem.target = self;
        [self.trayMenu addItem:toggleItem];

        NSMenuItem *settingsItem = [[NSMenuItem alloc] initWithTitle:@"偏好设置..." action:@selector(onMenuSettings:) keyEquivalent:@","];
        settingsItem.target = self;
        [self.trayMenu addItem:settingsItem];

        [self.trayMenu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"退出 GoClip" action:@selector(onMenuQuit:) keyEquivalent:@"q"];
        quitItem.target = self;
        [self.trayMenu addItem:quitItem];
    });
}

- (void)onButtonClick:(id)sender {
    NSEvent *event = [NSApp currentEvent];
    if (event.type == NSEventTypeRightMouseUp) {
        // 右键弹出菜单
        [self.statusItem popUpStatusItemMenu:self.trayMenu];
    } else {
        // 左键点击直接唤出/隐藏主窗口
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
}

@end

static GoClipTrayController *globalTray = nil;

void initMacTray(const void *bytes, int len) {
    if (globalTray == nil) {
        globalTray = [[GoClipTrayController alloc] init];
        [globalTray setupTrayWithBytes:bytes length:len];
    }
}
