#include "ApoCliUrlHandler.h"

@implementation BrowseAppDelegate
- (void)applicationWillFinishLaunching:(NSNotification *)aNotification {
    NSAppleEventManager *appleEventManager = [NSAppleEventManager sharedAppleEventManager];
    [appleEventManager setEventHandler:self
        andSelector:@selector(handleGetURLEvent:withReplyEvent:)
        forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender {
    return NSTerminateNow;
}

- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
    HandleURL((char*)[[[event paramDescriptorForKeyword:keyDirectObject] stringValue] UTF8String]);
}
@end

void RunApp(void) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setDelegate:[[BrowseAppDelegate alloc] init]];
        [NSApp run];
    }
}

int ShowAlert(bool error, char *message, char *details) {
	NSBundle *bundle = [NSBundle mainBundle];
	NSString *icon_path = nil;
	NSURL *icon_url = nil;
	NSString *icon_file = [bundle objectForInfoDictionaryKey:@"CFBundleIconFile"];
	if (icon_file != nil) {
		icon_path = [[[bundle resourcePath] stringByAppendingString:@"/"] stringByAppendingString:icon_file];
		if ([[icon_path pathExtension] length] == 0) icon_path = [icon_path stringByAppendingPathExtension:@"icns"];
		icon_url = [NSURL URLWithString:icon_path];
	}

    CFOptionFlags cfRes;

    CFUserNotificationDisplayAlert(
        0,
        error ? kCFUserNotificationStopAlertLevel : kCFUserNotificationPlainAlertLevel,
        (CFURLRef)icon_url,
        NULL,
        NULL,
        (__bridge CFStringRef)icon_file,//CFStringCreateWithCString(NULL, message, kCFStringEncodingUTF8),
        CFStringCreateWithCString(NULL, details, kCFStringEncodingUTF8),
        CFSTR("OK"),
        error ? NULL : CFSTR("Cancel"),
        NULL,
        &cfRes);

    return cfRes;
}