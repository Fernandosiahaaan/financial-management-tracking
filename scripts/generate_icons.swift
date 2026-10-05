import Cocoa

func renderIcon(size: CGFloat, path: String) {
    let rep = NSBitmapImageRep(
        bitmapDataPlanes: nil,
        pixelsWide: Int(size),
        pixelsHigh: Int(size),
        bitsPerSample: 8,
        samplesPerPixel: 4,
        hasAlpha: true,
        isPlanar: false,
        colorSpaceName: .deviceRGB,
        bytesPerRow: 0,
        bitsPerPixel: 0
    )!
    
    NSGraphicsContext.saveGraphicsState()
    let context = NSGraphicsContext(bitmapImageRep: rep)!
    NSGraphicsContext.current = context
    
    let rect = CGRect(x: 0, y: 0, width: size, height: size)
    
    // Background gradient
    let bgPath = NSBezierPath(rect: rect)
    let startColor = NSColor(calibratedRed: 79/255.0, green: 70/255.0, blue: 229/255.0, alpha: 1.0) // Indigo-600
    let endColor = NSColor(calibratedRed: 30/255.0, green: 27/255.0, blue: 75/255.0, alpha: 1.0)   // Indigo-950
    let gradient = NSGradient(starting: startColor, ending: endColor)!
    gradient.draw(in: bgPath, angle: -45)
    
    // Glowing center circle
    let circleRadius = size * 0.38
    let circleRect = CGRect(x: (size - circleRadius * 2) / 2, y: (size - circleRadius * 2) / 2, width: circleRadius * 2, height: circleRadius * 2)
    let circlePath = NSBezierPath(ovalIn: circleRect)
    let glowColor = NSColor(calibratedRed: 99/255.0, green: 102/255.0, blue: 241/255.0, alpha: 0.35)
    glowColor.setFill()
    circlePath.fill()
    
    // Inner golden coin ring
    let coinRadius = size * 0.30
    let coinRect = CGRect(x: (size - coinRadius * 2) / 2, y: (size - coinRadius * 2) / 2, width: coinRadius * 2, height: coinRadius * 2)
    let coinPath = NSBezierPath(ovalIn: coinRect)
    let coinGrad = NSGradient(
        starting: NSColor(calibratedRed: 245/255.0, green: 158/255.0, blue: 11/255.0, alpha: 1.0),
        ending: NSColor(calibratedRed: 251/255.0, green: 191/255.0, blue: 36/255.0, alpha: 1.0)
    )!
    coinGrad.draw(in: coinPath, angle: 45)
    
    // Draw "FT" emblem inside coin
    let text = "FT"
    let fontSize = size * 0.28
    let font = NSFont.systemFont(ofSize: fontSize, weight: .black)
    let paragraphStyle = NSMutableParagraphStyle()
    paragraphStyle.alignment = .center
    let attrs: [NSAttributedString.Key: Any] = [
        .font: font,
        .foregroundColor: NSColor.white,
        .paragraphStyle: paragraphStyle
    ]
    let str = NSAttributedString(string: text, attributes: attrs)
    let textSize = str.size()
    let textRect = CGRect(
        x: (size - textSize.width) / 2,
        y: (size - textSize.height) / 2,
        width: textSize.width,
        height: textSize.height
    )
    str.draw(in: textRect)
    
    NSGraphicsContext.restoreGraphicsState()
    
    let pngData = rep.representation(using: .png, properties: [:])!
    try! pngData.write(to: URL(fileURLWithPath: path))
    print("Generated \(path) (\(Int(size))x\(Int(size)))")
}

let publicDir = "/Users/fernando/Documents/personal/repo/git/financial-management-tracking/frontend/public"
renderIcon(size: 180, path: "\(publicDir)/apple-touch-icon.png")
renderIcon(size: 192, path: "\(publicDir)/icon-192.png")
renderIcon(size: 512, path: "\(publicDir)/icon-512.png")
