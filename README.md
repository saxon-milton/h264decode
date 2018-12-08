# NightLife

Using an 840nm infrared light, a Rpi3.B+, and NoIR camera, find bright glowing orbs (animal eyes), humans without lights, and humans with lights. 

# Motivation

On occassion, the area around where I live has car break-ins and gas syphoning. We also have a diverse collection of wildlife varying from semi-domestic cats to coyotes and mountain lions.

## Detecting Cars

It seems like cars are going to create large area of brightness relative to the image size. We'll peg an image as a car if that's what is going on and keep tracking of +/- 2 seconds around its appearance.

## Detecting Animals

Night animals have a slight reflective coating to their eyes, this would be one factor in finding. Another factor is their height which would be likely 2/3 or less of what's determined as a human's height.

## Detecting Humans

Humans come with and without flashlights. Humans with flashlights are kind of like slow moving cars. Humans without lflashlights are tall vertical animals.

## Camera Notes

* AWB gains are in the (315/256, 90/64) in the late afternoon at the North side of the house (15:00 2018/12/07)
* Framerate of 40 seems stable
* AWB gains are in the (197/128, 163/128) range in the early twilight at the North side of the house (16:30 2018/12/07)


