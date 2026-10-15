//go:build darwin

#include <stdint.h>
#import <Carbon/Carbon.h>

extern void goHotkeyTriggered();

static OSStatus hotKeyHandler(EventHandlerCallRef nextHandler, EventRef anEvent, void *userData) {
    EventHotKeyID hkID;
    GetEventParameter(anEvent, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(hkID), NULL, &hkID);
    if (hkID.signature == 'GCLP') {
        goHotkeyTriggered();
    }
    return noErr;
}

static EventHotKeyRef currentHotKeyRef = NULL;
static int isHandlerInstalled = 0;

int registerCarbonHotKey(uint32_t keyCode, uint32_t modifiers) {
    if (!isHandlerInstalled) {
        EventTypeSpec eventType;
        eventType.eventClass = kEventClassKeyboard;
        eventType.eventKind = kEventHotKeyPressed;
        InstallApplicationEventHandler(&hotKeyHandler, 1, &eventType, NULL, NULL);
        isHandlerInstalled = 1;
    }

    if (currentHotKeyRef != NULL) {
        UnregisterEventHotKey(currentHotKeyRef);
        currentHotKeyRef = NULL;
    }

    EventHotKeyID hkID;
    hkID.signature = 'GCLP';
    hkID.id = 1;

    OSStatus status = RegisterEventHotKey(keyCode, modifiers, hkID, GetApplicationEventTarget(), 0, &currentHotKeyRef);
    return (status == noErr) ? 0 : (int)status;
}

void unregisterCarbonHotKey() {
    if (currentHotKeyRef != NULL) {
        UnregisterEventHotKey(currentHotKeyRef);
        currentHotKeyRef = NULL;
    }
}
